# List Members Available to Add to Room

## Goal
Change `POST /api/members` from listing all members to listing members who have NOT yet joined a given chat room.

## Changes

### 1. Migration (`db/migrations/011_list_members.up.sql`)
- Replace `list_members`/`count_members` with `list_members_not_in_room(p_room_id, p_limit, p_offset)` and `count_members_not_in_room(p_room_id)`

### 2. Types
- `src/domain/chat.go`: Add `RoomMembersRequest` with `RoomID`, `Page`, `Limit`
- `src/domain/member.go`: Remove `ListMembersRequest`

### 3. Repository interface (`src/domain/repository.go`)
- Remove `ListMembers(ctx, limit, offset)` and `CountMembers(ctx)` from `MemberRepository`
- Add `ListMembersNotInRoom(ctx, roomID, limit, offset)` and `CountMembersNotInRoom(ctx, roomID)` to `ChatRoomRepository`

### 4. Repository impl (`src/database/chat_repo.go`)
- Add `memberRow` scan struct, `ListMembersNotInRoom` and `CountMembersNotInRoom` methods on `ChatRoomRepositoryPGX`

### 5. Handler (`src/api/chat_handler.go`)
- `ChatRoomHandler` has no `memberRepo` field; uses `h.repo.ListMembersNotInRoom`
- `ListAvailableMembers` handler, `roomID` from `mux.Vars(r)["roomId"]`

### 6. Routes (`src/api/router.go`)
- `POST /api/chat/rooms/{roomId}/available-members` in `RegisterChatRoutes`
- Removed from `RegisterMemberRoutes`

### 7. Tests (`src/test/api/chat_handler_test.go`)
- `TestChatHandler_ListAvailableMembers`: list available members + unauthorized
