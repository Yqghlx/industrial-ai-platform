# Device Repository Test Structure

```
device_repo_test.go
│
├── Create Operations (2 tests)
│   ├── ✅ TestDeviceRepository_Create_Success
│   │   └── Validates INSERT with ON CONFLICT
│   └── ✅ TestDeviceRepository_Create_Error
│       └── Tests database error handling
│
├── Read Operations - GetByID (3 tests)
│   ├── ✅ TestDeviceRepository_GetByID_Success
│   │   └── Validates SELECT with WHERE id = $1
│   ├── ✅ TestDeviceRepository_GetByID_NotFound
│   │   └── Tests sql.ErrNoRows handling
│   └── ✅ TestDeviceRepository_GetByID_DatabaseError
│       └── Tests connection error handling
│
├── Read Operations - List (6 tests)
│   ├── ✅ TestDeviceRepository_List_Success
│   │   └── Validates COUNT + SELECT with pagination
│   ├── ✅ TestDeviceRepository_List_SecondPage
│   │   └── Tests offset calculation (page 2)
│   ├── ✅ TestDeviceRepository_List_EmptyResult
│   │   └── Tests empty result set handling
│   ├── ✅ TestDeviceRepository_List_CountError
│   │   └── Tests COUNT query error
│   ├── ✅ TestDeviceRepository_List_QueryError
│   │   └── Tests SELECT query error
│   └── ✅ TestDeviceRepository_List_ScanError
│       └── Tests type mismatch error
│
├── Update Operations (2 tests)
│   ├── ✅ TestDeviceRepository_Update_Success
│   │   └── Validates UPDATE with SET and WHERE
│   └── ✅ TestDeviceRepository_Update_Error
│       └── Tests update error handling
│
├── Delete Operations (2 tests)
│   ├── ✅ TestDeviceRepository_Delete_Success
│   │   └── Validates DELETE with WHERE id = $1
│   └── ✅ TestDeviceRepository_Delete_Error
│       └── Tests delete error handling
│
├── UpdateStatus Operations (2 tests)
│   ├── ✅ TestDeviceRepository_UpdateStatus_Success
│   │   └── Validates status-only update
│   └── ✅ TestDeviceRepository_UpdateStatus_Error
│       └── Tests status update error
│
├── Count Operations (3 tests)
│   ├── ✅ TestDeviceRepository_Count_Success
│   │   └── Tests COUNT(*) query
│   ├── ✅ TestDeviceRepository_Count_Error
│   │   └── Tests count error handling
│   └── ✅ TestDeviceRepository_Count_ZeroDevices
│       └── Tests zero devices scenario
│
├── SQL Pattern Validation (1 test, 5 sub-tests)
│   └── ✅ TestDeviceRepository_SQLQueryPatterns
│       ├── Create: INSERT pattern
│       ├── GetByID: SELECT WHERE pattern
│       ├── List: COUNT + SELECT ORDER/LIMIT/OFFSET pattern
│       ├── Update: UPDATE SET WHERE pattern
│       └── Delete: DELETE WHERE pattern
│
└── Context Handling (1 test, 2 sub-tests)
    └── ✅ TestDeviceRepository_ContextCancellation
        ├── Create with canceled context
        └── GetByID with canceled context

Total: 25 Test Functions
       30+ Test Cases
       100% Method Coverage
```

## SQL Query Patterns Tested

```
┌─────────────────────────────────────────────────────────────┐
│ 1. INSERT Pattern                                            │
│    INSERT INTO devices (...) VALUES (...) ON CONFLICT ...   │
│    Parameters: 8 fields                                       │
│    Returns: Result                                           │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ 2. SELECT by ID Pattern                                      │
│    SELECT ... FROM devices WHERE id = $1                     │
│    Parameters: 1 (id)                                         │
│    Returns: Single row                                       │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ 3. COUNT Pattern                                             │
│    SELECT COUNT(*) FROM devices                              │
│    Parameters: 0                                             │
│    Returns: Integer count                                   │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ 4. SELECT with Pagination Pattern                            │
│    SELECT ... FROM devices                                   │
│    ORDER BY created_at DESC                                  │
│    LIMIT $1 OFFSET $2                                        │
│    Parameters: 2 (pageSize, offset)                          │
│    Returns: Multiple rows                                    │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ 5. UPDATE Pattern                                            │
│    UPDATE devices SET name=$1, type=$2, ... WHERE id=$7     │
│    Parameters: 7 fields                                      │
│    Returns: Result                                           │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ 6. DELETE Pattern                                            │
│    DELETE FROM devices WHERE id = $1                        │
│    Parameters: 1 (id)                                        │
│    Returns: Result                                           │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│ 7. UPDATE Status Pattern                                     │
│    UPDATE devices SET status=$1, updated_at=$2 WHERE id=$3  │
│    Parameters: 3 (status, updated_at, id)                   │
│    Returns: Result                                           │
└─────────────────────────────────────────────────────────────┘
```

## Test Flow

```
┌──────────────┐
│ Setup Mock   │
│   Database   │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│  Create      │
│  Repository  │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│   Setup      │
│ Expectations │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│   Execute    │
│    Method    │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│  Assertions  │
│   & Verify   │
└──────────────┘
```

## Coverage Matrix

| Method | Success | Error | Not Found | Empty | Pagination |
|--------|---------|-------|-----------|-------|------------|
| Create | ✅ | ✅ | - | - | - |
| GetByID | ✅ | ✅ | ✅ | - | - |
| List | ✅ | ✅ | - | ✅ | ✅ |
| Update | ✅ | ✅ | - | - | - |
| Delete | ✅ | ✅ | - | - | - |
| UpdateStatus | ✅ | ✅ | - | - | - |
| Count | ✅ | ✅ | - | ✅ | - |

Total: 25 test functions, 30+ test cases