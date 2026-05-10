# Tinder Matching System (In-Memory)

這是一個基於 Go 語言實作的高效能單身配對 HTTP 伺服器，專為處理高併發場景設計，並採用純記憶體儲存（In-Memory Storage）。

## 目錄
* [核心技術亮點](#核心技術亮點)
* [專案架構](#專案架構)
* [快速啟動](#快速啟動)
* [API 概覽](#api-概覽)
* [效能複雜度分析](#效能複雜度分析)
* [TBD (待優化事項)](#tbd-待優化事項)

---
## 核心技術亮點

*   **極速搜尋效能 O(S_{opp})**：採用 **性別分桶** ，將搜尋範圍精準鎖定在異性集合，搜尋效率提升。
*   **併發安全機制**：
    *   **分段鎖 (Lock Stripping)**：針對男/女池配置獨立的 `RWMutex`，降低鎖競爭，支援高並發存取。
    *   **原子化兌換 (Atomic Redemption)**：實作 `RedeemWantedDates` 原子操作，確保配對次數在極高併發下精準扣除，杜絕遺失更新 (Lost Update) 問題。
*   **寫入旁路優化 (Write-Bypass)**：若使用者進場即配對完成，系統將跳過入庫流程，節省記憶體開銷。
*   **生產級日誌監控**：整合結構化日誌 (`slog`) 與 Middleware，自動紀錄請求耗時與 `error_code`。

---

## 專案架構

```text
├── cmd/server/main.go         # 配置 + 日誌 + 容器初始化
├── internal/
│   ├── config/                # 運行配置
│   ├── container/             # (DI Container)
│   ├── gateway/               # HTTP 傳輸層 (Handlers/Router/Middleware)
│   ├── model/                 # 請求/回應結構體
│   ├── repository/            # 存儲層介面與實作
│   └── service/               # 業務邏輯介面與實作
└── docs/                      # 系統設計與 API 詳細文件
```

---

## 快速啟動

### 1. 執行伺服器
```bash
go run ./cmd/server/main.go
```
伺服器將預設運行在 `http://localhost:8080`。

### 2. 運行單元測試與覆蓋率
```bash
# 執行全專案測試
go test -v ./...

# 查看特定模組覆蓋率
go test -cover ./internal/service/...
go test -cover ./internal/gateway/...
```

---

## API 概覽

詳細資訊請參考 [API Documentation](./docs/API.md)。


| 方法 | 路徑 | 功能 |
| :--- | :--- | :--- |
| `POST` | `/api/v1/people/match` | 新增單身人士並立即執行配對 |
| `DELETE` | `/api/v1/people/{name}` | 從系統中移除特定成員 |
| `GET` | `/api/v1/people` | 列出所有目前單身成員 (姓名排序) |
| `GET` | `/api/v1/people/{name}` | 查詢特定人員詳細資料 |
| `GET` | `/api/v1/people/{name}/matches?top={N}` | 查詢符合該人員條件的 Top N 候選人 |

---

## 效能複雜度分析


| API 功能 | URL | API Name | 複雜度 | 說明 |
| :--- | :--- | :--- | :--- | :--- |
| **新增並配對** | `POST /api/v1/people/match` | `handleAddSinglePersonAndMatch` | O(S_opp) | 包含 O(1) 唯一性檢查與異性池一趟掃描 (Single Pass)。 |
| **移除人員** | `DELETE /api/v1/people/{name}` | `handleRemoveSinglePerson` | O(1) | 分區 Map 雜湊檢索，常數時間定位。 |
| **查詢配對名單** | `GET /api/v1/people/{name}/matches` | `handleQueryPersonMatches` | O(S_opp + K log K) | 異性池篩選 O(S_opp) 與結果排序 O(K log K)。 |
| **查詢單一成員** | `GET /api/v1/people/{name}` | `handleQuerySinglePerson` | O(1) | 透過名稱雜湊直接檢索。 |
| **查詢所有成員** | `GET /api/v1/people` | `handleQuerySinglePeople` | O(S log S) | 全量資料提取 O(S) 並進行穩定排序。 |

*詳細設計邏輯與複雜度推導請參閱 [System Design](./docs/DESIGN.md)。*

---

## TBD (待優化事項)

- **唯一值設計優化**: 
  - 目前以 `name` 作為唯一鍵（Primary Key）不符合現實場景（同名問題）。
  - 計畫改用 **UUID** 或 **ULID** 作為內部唯一識別碼，提升資料可靠性與隱私保護。
- **分頁處理**: 
  - 為 `QuerySinglePeople` 增加 `limit` 與 `offset` (或 Cursor-based) 分頁支援，避免大數據量下單次 Response 過大。
- **搜尋索引優化**: 
  - 建立 **多維度配對權重機制** (如地理位置、興趣標籤)，提升配對精準度。
- **即時通知系統 (Real-time Notification)**:
  - 實作 **WebSocket** 或 **SSE (Server-Sent Events)** 雙向/單向通訊機制。
  - 當新使用者註冊並觸發配對時，系統能主動推播 (Push Notification) 給被配對到的在線使用者