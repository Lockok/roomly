import { FormEvent, useEffect, useMemo, useState } from 'react'
import { CalendarDays, Check, Edit3, Eye, EyeOff, LogOut, MapPin, Plus, Power, Save, Users, X } from 'lucide-react'

const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

type Room = {
  id: string
  name: string
  location: string
  floor?: number
  capacity: number
  equipment: string[]
  description?: string
  is_active: boolean
}

type Booking = {
  id: string
  room_id: string
  title: string
  starts_at: string
  ends_at: string
  status: 'confirmed' | 'cancelled'
  attendees_count: number
}

type ApiError = { message?: string; error?: { message?: string } }

function toInputDate(date: Date) {
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60000)
  return local.toISOString().slice(0, 16)
}

function initialWindow() {
  const start = new Date()
  start.setMinutes(0, 0, 0)
  start.setHours(start.getHours() + 1)
  const end = new Date(start)
  end.setHours(end.getHours() + 1)
  return { start: toInputDate(start), end: toInputDate(end) }
}

async function request<T>(path: string, options: RequestInit = {}, token?: string): Promise<T> {
  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(options.headers ?? {}),
    },
  })

  if (!response.ok) {
    const body = (await response.json().catch(() => ({}))) as ApiError
    throw new Error(body.message ?? body.error?.message ?? `Ошибка запроса (${response.status})`)
  }

  return response.json() as Promise<T>
}

function decodeClaims(token: string) {
  try {
    return JSON.parse(atob(token.split('.')[1])) as { user_id?: string; role?: string }
  } catch {
    return {}
  }
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat('ru-RU', { weekday: 'short', day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' }).format(new Date(value))
}

function App() {
  const [token, setToken] = useState(() => localStorage.getItem('roomly_token') ?? '')
  const [authMode, setAuthMode] = useState<'login' | 'register'>('login')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [fullName, setFullName] = useState('')
  const [loginError, setLoginError] = useState('')

  if (!token) {
    return (
      <main className="auth-shell">
        <section className="auth-copy">
          <span className="brand-mark">ROOMLY / 01</span>
          <h1>Встречи, которые начинаются вовремя.</h1>
          <p>Единое пространство для переговорных, расписаний и спокойного рабочего дня.</p>
          <div className="auth-note"><CalendarDays size={18} /> <span>Актуальная занятость без пересечений</span></div>
        </section>
        <form className="auth-form" onSubmit={async (event) => {
          event.preventDefault()
          setLoginError('')
          try {
            if (authMode === 'register') {
              await request('/api/v1/users', {
                method: 'POST', body: JSON.stringify({ email, password, full_name: fullName }),
              })
              setAuthMode('login')
              setLoginError('Аккаунт создан. Теперь войдите с вашим паролем.')
              return
            }

            const result = await request<{ access_token: string }>('/api/v1/auth/login', {
              method: 'POST', body: JSON.stringify({ email, password }),
            })
            localStorage.setItem('roomly_token', result.access_token)
            setToken(result.access_token)
          } catch (error) {
            setLoginError(error instanceof Error ? error.message : 'Не удалось войти')
          }
        }}>
          <p className="eyebrow">Рабочее пространство</p>
          <h2>{authMode === 'login' ? 'Войти в Roomly' : 'Создать аккаунт'}</h2>
          {authMode === 'register' && <label>Имя<input value={fullName} onChange={(event) => setFullName(event.target.value)} required placeholder="Ваше имя" /></label>}
          <label>Email<input type="email" value={email} onChange={(event) => setEmail(event.target.value)} required placeholder="you@company.com" /></label>
          <label className="password-field">Пароль<div className="password-input"><input type={showPassword ? 'text' : 'password'} minLength={8} value={password} onChange={(event) => setPassword(event.target.value)} required placeholder="Минимум 8 символов" /><button className="password-toggle" type="button" onClick={() => setShowPassword((visible) => !visible)} aria-label={showPassword ? 'Скрыть пароль' : 'Показать пароль'} title={showPassword ? 'Скрыть пароль' : 'Показать пароль'}>{showPassword ? <EyeOff size={18} /> : <Eye size={18} />}</button></div></label>
          {loginError && <p className={authMode === 'register' && loginError.startsWith('Аккаунт') ? 'success-text' : 'error-text'}>{loginError}</p>}
          <button className="primary-button" type="submit">{authMode === 'login' ? 'Открыть расписание' : 'Зарегистрироваться'} <span>↗</span></button>
          <button className="auth-switch" type="button" onClick={() => { setAuthMode(authMode === 'login' ? 'register' : 'login'); setLoginError('') }}>
            {authMode === 'login' ? 'Нет аккаунта? Зарегистрироваться' : 'Уже есть аккаунт? Войти'}
          </button>
        </form>
      </main>
    )
  }

  return <Dashboard token={token} onLogout={() => { localStorage.removeItem('roomly_token'); setToken('') }} />
}

function Dashboard({ token, onLogout }: { token: string; onLogout: () => void }) {
  const windowDefaults = useMemo(initialWindow, [])
  const claims = useMemo(() => decodeClaims(token), [token])
  const isAdmin = claims.role === 'admin'
  const [rooms, setRooms] = useState<Room[]>([])
  const [bookings, setBookings] = useState<Booking[]>([])
  const [startsAt, setStartsAt] = useState(windowDefaults.start)
  const [endsAt, setEndsAt] = useState(windowDefaults.end)
  const [selectedRoom, setSelectedRoom] = useState('')
  const [title, setTitle] = useState('')
  const [attendees, setAttendees] = useState(2)
  const [status, setStatus] = useState('')
  const [error, setError] = useState('')

  async function loadRooms() {
    try {
      const query = new URLSearchParams({ starts_at: new Date(startsAt).toISOString(), ends_at: new Date(endsAt).toISOString() })
      const result = await request<Room[]>(`/api/v1/rooms/available?${query}`, {}, token)
      setRooms(result)
      setSelectedRoom((current) => result.some((room) => room.id === current) ? current : result[0]?.id ?? '')
      setError('')
    } catch (loadError) { setError(loadError instanceof Error ? loadError.message : 'Не удалось загрузить комнаты') }
  }

  async function loadBookings() {
    const userId = claims.user_id ?? ''
    const query = userId ? `?organizer_id=${userId}` : ''
    try {
      const result = await request<{ items: Booking[] }>(`/api/v1/bookings${query}`, {}, token)
      setBookings(result.items)
    } catch (loadError) { setError(loadError instanceof Error ? loadError.message : 'Не удалось загрузить встречи') }
  }

  useEffect(() => { void loadRooms(); void loadBookings() }, [])

  async function createBooking(event: FormEvent) {
    event.preventDefault()
    setStatus('')
    setError('')
    try {
      await request('/api/v1/bookings', { method: 'POST', body: JSON.stringify({ room_id: selectedRoom, title, starts_at: new Date(startsAt).toISOString(), ends_at: new Date(endsAt).toISOString(), attendees_count: attendees }) }, token)
      setTitle('')
      setStatus('Встреча создана')
      await Promise.all([loadRooms(), loadBookings()])
    } catch (createError) { setError(createError instanceof Error ? createError.message : 'Не удалось создать встречу') }
  }

  async function cancelBooking(id: string) {
    try {
      await request(`/api/v1/bookings/${id}/cancel`, { method: 'POST' }, token)
      await loadBookings()
    } catch (cancelError) { setError(cancelError instanceof Error ? cancelError.message : 'Не удалось отменить встречу') }
  }

  return (
    <div className="app-shell">
      <header className="topbar"><div className="brand-mark">ROOMLY / 01</div><div className="topbar-actions"><span className="status-dot">{isAdmin ? 'Администратор' : 'Система в норме'}</span><button className="icon-button" onClick={onLogout} title="Выйти"><LogOut size={18} /></button></div></header>
      <main className="dashboard">
        <section className="page-heading"><div><p className="eyebrow">Панель бронирований</p><h1>Найдите место для следующей идеи.</h1></div><div className="date-stamp">{new Intl.DateTimeFormat('ru-RU', { weekday: 'long', day: 'numeric', month: 'long' }).format(new Date())}</div></section>
        <section className="availability-panel">
          <div className="panel-heading"><div><span className="section-index">01</span><h2>Найти свободную комнату</h2></div><button className="text-button" onClick={() => { void loadRooms() }}>Обновить доступность ↗</button></div>
          <div className="filters"><label>Начало<input type="datetime-local" value={startsAt} onChange={(event) => setStartsAt(event.target.value)} /></label><label>Окончание<input type="datetime-local" value={endsAt} onChange={(event) => setEndsAt(event.target.value)} /></label><button className="primary-button filter-button" onClick={() => { void loadRooms() }}>Показать комнаты</button></div>
          <div className="room-grid">{rooms.map((room) => <button className={`room-card ${selectedRoom === room.id ? 'selected' : ''}`} key={room.id} onClick={() => setSelectedRoom(room.id)}><div className="room-card-top"><span className="room-number">{room.name.slice(0, 2).toUpperCase()}</span>{selectedRoom === room.id && <Check size={18} />}</div><h3>{room.name}</h3><p><MapPin size={14} /> {room.location}{room.floor ? ` · этаж ${room.floor}` : ''}</p><div className="room-meta"><span><Users size={14} /> до {room.capacity}</span><span>{room.equipment.slice(0, 2).join(' · ')}</span></div></button>)}{rooms.length === 0 && <div className="empty-state">На этот интервал свободных комнат нет.</div>}</div>
        </section>
        {isAdmin && <AdminRooms token={token} onError={setError} onStatus={setStatus} />}
        <div className="content-grid">
          <section className="booking-panel"><div className="panel-heading"><div><span className="section-index">02</span><h2>Забронировать</h2></div></div><form onSubmit={createBooking}><label>Название встречи<input value={title} onChange={(event) => setTitle(event.target.value)} required placeholder="Например, планирование квартала" /></label><label>Участники<input type="number" min="1" max="1000" value={attendees} onChange={(event) => setAttendees(Number(event.target.value))} required /></label><button className="primary-button" disabled={!selectedRoom} type="submit">Забронировать комнату <span>↗</span></button>{status && <p className="success-text">{status}</p>}</form></section>
          <section className="agenda-panel"><div className="panel-heading"><div><span className="section-index">03</span><h2>Ваши встречи</h2></div><span className="booking-count">{bookings.filter((booking) => booking.status === 'confirmed').length} активных</span></div><div className="agenda-list">{bookings.filter((booking) => booking.status === 'confirmed').map((booking) => <article className="agenda-item" key={booking.id}><div className="agenda-time">{formatDate(booking.starts_at)}<br /><span>до {new Intl.DateTimeFormat('ru-RU', { hour: '2-digit', minute: '2-digit' }).format(new Date(booking.ends_at))}</span></div><div><h3>{booking.title}</h3><p>{booking.attendees_count} участника · {rooms.find((room) => room.id === booking.room_id)?.name ?? 'Комната'}</p></div><button className="icon-button danger" onClick={() => { void cancelBooking(booking.id) }} title="Отменить"><X size={17} /></button></article>)}{bookings.length === 0 && <div className="empty-state">Запланированных встреч пока нет.</div>}</div></section>
        </div>
        {error && <div className="toast error-text">{error}</div>}
      </main>
    </div>
  )
}

function AdminRooms({ token, onError, onStatus }: { token: string; onError: (message: string) => void; onStatus: (message: string) => void }) {
  const [rooms, setRooms] = useState<Room[]>([])
  const [loading, setLoading] = useState(true)
  const [editingRoomId, setEditingRoomId] = useState<string | null>(null)
  const [editingRoomName, setEditingRoomName] = useState('')
  const [editingRoomCapacity, setEditingRoomCapacity] = useState(1)

  async function loadAllRooms() {
    try {
      setLoading(true)
      setRooms(await request<Room[]>('/api/v1/rooms', {}, token))
    } catch (loadError) {
      onError(loadError instanceof Error ? loadError.message : 'Не удалось загрузить комнаты')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void loadAllRooms() }, [])

  async function toggleRoom(room: Room) {
    try {
      const updated = await request<Room>(`/api/v1/rooms/${room.id}`, {
        method: 'PATCH',
        body: JSON.stringify({ is_active: !room.is_active }),
      }, token)
      setRooms((current) => current.map((item) => item.id === updated.id ? updated : item))
      onStatus(updated.is_active ? 'Комната снова доступна' : 'Комната отключена')
    } catch (updateError) {
      onError(updateError instanceof Error ? updateError.message : 'Не удалось изменить комнату')
    }
  }

  return <section className="admin-panel">
    <div className="panel-heading"><div><span className="section-index">ADMIN</span><h2>Управление комнатами</h2></div><span className="booking-count">{rooms.length} всего</span></div>
    <form className="admin-create-form" onSubmit={async (event) => {
      event.preventDefault()
      const formElement = event.currentTarget
      const form = new FormData(formElement)
      try {
        await request('/api/v1/rooms', {
          method: 'POST',
          body: JSON.stringify({ name: form.get('name'), location: form.get('location'), capacity: Number(form.get('capacity')), equipment: String(form.get('equipment') ?? '').split(',').map((item) => item.trim()).filter(Boolean) }),
        }, token)
        formElement.reset()
        onStatus('Комната создана')
        await loadAllRooms()
      } catch (createError) { onError(createError instanceof Error ? createError.message : 'Не удалось создать комнату') }
    }}>
      <input name="name" required placeholder="Название комнаты" />
      <input name="location" required placeholder="Локация" />
      <input name="capacity" required min="1" type="number" placeholder="Мест" />
      <input name="equipment" placeholder="Оборудование через запятую" />
      <button className="primary-button" type="submit"><Plus size={17} /> Создать комнату</button>
    </form>
    <div className="admin-room-list">{loading ? <p className="empty-state">Загрузка комнат...</p> : rooms.map((room) => <div className="admin-room-row" key={room.id}>
      <div>{editingRoomId === room.id ? <div className="admin-edit-fields"><input className="admin-name-input" value={editingRoomName} onChange={(event) => setEditingRoomName(event.target.value)} aria-label="Название комнаты" /><input className="admin-capacity-input" type="number" min="1" value={editingRoomCapacity} onChange={(event) => setEditingRoomCapacity(Number(event.target.value))} aria-label="Количество мест" /></div> : <strong>{room.name}</strong>}<span>{room.location} · до {room.capacity} мест</span></div>
      <div className="admin-room-actions">{editingRoomId === room.id ? <button className="availability-button active" onClick={async () => {
        try {
          if (!editingRoomName.trim() || editingRoomCapacity < 1) {
            onError('Название комнаты и количество мест обязательны')
            return
          }
          const updated = await request<Room>(`/api/v1/rooms/${room.id}`, { method: 'PATCH', body: JSON.stringify({ name: editingRoomName.trim(), capacity: editingRoomCapacity }) }, token)
          setRooms((current) => current.map((item) => item.id === updated.id ? updated : item))
          setEditingRoomId(null)
          onStatus('Данные комнаты изменены')
        } catch (updateError) { onError(updateError instanceof Error ? updateError.message : 'Не удалось изменить комнату') }
      }}><Save size={15} /> Сохранить</button> : <button className="availability-button" onClick={() => { setEditingRoomId(room.id); setEditingRoomName(room.name); setEditingRoomCapacity(room.capacity) }}><Edit3 size={15} /> Изменить</button>}
        <button className={`availability-button ${room.is_active ? 'active' : ''}`} onClick={() => { void toggleRoom(room) }}><Power size={15} /> {room.is_active ? 'Доступна' : 'Отключена'}</button>
      </div>
    </div>)}</div>
  </section>
}

export default App
