# Events Feature

## Goal
Create a calendar event system with visibility control. Events appear on a frontend calendar, and creators can grant viewing access to other members.

## Tables

### events
| Column      | Type        | Notes                                |
|-------------|-------------|--------------------------------------|
| id          | UUID        | PK, gen_random_uuid()                |
| title       | VARCHAR(255)| NOT NULL                             |
| description | TEXT        | DEFAULT ''                           |
| start_time  | TIMESTAMPTZ | NOT NULL                             |
| end_time    | TIMESTAMPTZ | NOT NULL                             |
| status      | VARCHAR(50) | DEFAULT 'active'                     |
| created_by  | UUID        | FK -> members(id)                    |
| created_at  | TIMESTAMPTZ | DEFAULT now()                        |
| updated_at  | TIMESTAMPTZ | DEFAULT now()                        |

### event_viewers
| Column    | Type        | Notes                                         |
|-----------|-------------|-----------------------------------------------|
| event_id  | UUID        | PK, FK -> events(id) ON DELETE CASCADE        |
| member_id | UUID        | PK, FK -> members(id) ON DELETE CASCADE       |
| created_at| TIMESTAMPTZ | DEFAULT now()                                 |

## Stored Functions
1. `create_event(p_title, p_description, p_start_time, p_end_time, p_status, p_created_by)`
2. `get_event_by_id(p_id, p_member_id)` — checks visibility
3. `list_events_by_month(p_year, p_month, p_member_id)` — shows events user can see
4. `update_event(p_id, p_title, p_description, p_start_time, p_end_time, p_status)`
5. `delete_event(p_id)`
6. `add_event_viewer(p_event_id, p_member_id)`
7. `remove_event_viewer(p_event_id, p_member_id)`
8. `list_event_viewers(p_event_id)`

## API Endpoints (all auth'd)
| Method | Path | Purpose |
|--------|------|---------|
| GET    | /api/events?year=&month= | List visible events by month |
| POST   | /api/events | Create event |
| GET    | /api/events/{eventId} | Get event detail (if visible) |
| POST   | /api/events/{eventId}/update | Update event (owner only) |
| POST   | /api/events/{eventId}/delete | Delete event (owner only) |
| POST   | /api/events/{eventId}/viewers | Add viewer (owner only) |
| POST   | /api/events/{eventId}/viewers/{memberId}/remove | Remove viewer (owner only) |
| GET    | /api/events/{eventId}/viewers | List viewers (owner only) |

## Files
- **New:** `db/migrations/012_create_events.up.sql`
- **New:** `src/domain/event.go`
- **New:** `src/database/event_repo.go`
- **New:** `src/api/event_handler.go`
- **Modified:** `src/domain/repository.go` — add EventRepository interface
- **Modified:** `src/domain/errors.go` — add event errors
- **Modified:** `src/api/router.go` — add RegisterEventRoutes
- **Modified:** `main.go` — wire up routes
