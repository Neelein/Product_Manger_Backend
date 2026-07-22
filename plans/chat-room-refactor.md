# Chat Room Refactor: Remove Type Field

## Goal
Simplify chat rooms by removing the `type` (direct/group) distinction. Users create a room with just a name, then add members via the existing AddMembers endpoint.

## Changes

### 1. Domain (`src/domain/chat.go`)
- Remove `Type` field from `ChatRoom` struct
- Remove `Type` from `ChatRoomWithMeta` (embedded)
- Simplify `CreateRoomRequest` to only have `Name`

### 2. Repository (`src/database/chat_repo.go`)
- Remove `Type` parameter from `CreateRoom` SQL call to `create_chat_room`
- Default to 'group' type in the function
- Add only creator as member after room creation

### 3. Handler (`src/api/chat_handler.go`)
- Remove type validation in `CreateRoom`
- Remove member_ids handling in `CreateRoom`
- Only require `name` in request

### 4. Migration SQL (`db/migrations/010_create_chat.up.sql`)
- Remove `type` column from `chat_rooms` table
- Remove `p_type` parameter from `create_chat_room` function
- Default type to 'group'

### 5. Tests
- Update all test references to `Type` field
- Update `CreateRoom` calls to not pass type
- Remove direct room tests
- Simplify `CreateRoomRequest` usage
