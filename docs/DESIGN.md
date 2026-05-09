# 系統設計文件 (System Design)

## 1. 系統目標 (Goal)
這是一個基於 Go 語言實作的高效能單身配對 HTTP 伺服器，專為處理高併發場景設計，並採用純記憶體儲存（In-Memory Storage）。

## 2. 專案架構 (Project Layout)

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

## 3. 核心設計亮點 (Design Highlights)

### 3.1 性別分桶 (Gender Partitioning)
Repository 層將資料預先拆分為 `males` 與 `females` 兩個獨立 Map。
- **效能優勢**：在配對搜尋時，搜尋範圍直接縮小至特定的異性集合 S_opp，同性別成員的搜尋複雜度為 O(0)。

### 3.2 分鎖 (Lock Stripping)
為男/女池配置獨立的 `sync.RWMutex`。
- **併發優勢**：當系統正在處理男性成員註冊時，不會阻塞女性成員池的讀寫，大幅降低鎖競爭 (Lock Contention)，提升系統吞吐量。

### 3.3 原子化兌換機制 (Atomic Redemption)
為解決高併發場景下「先讀後寫」導致的 **遺失更新 (Lost Update)** 問題，實作了 `RedeemWantedDates` 原子方法。
- **一致性保證**：將「檢查剩餘次數」與「執行扣除」封裝在單一寫鎖區間內，確保配對次數在高併發下精準扣除，避免資源超配。

### 3.4 寫入優化 
若使用者在進場執行新增配對時即耗盡配對次數，系統將直接回傳結果，跳過 Repository `Upsert` 流程，減少不必要的記憶體寫入開銷。

## 4. API 時間複雜度 (Time Complexity)

本系統採用性別分桶，將搜尋範圍從總人數 S (M + F) 縮小至特定的異性集合。
- **S_opp**: 異性集合總數 (若發起者為男則為 F，若為女則為 M)
- **K**: 符合特定條件（如身高）的對象人數 (K <= S_opp)


| API 功能 | URL | API Name | 複雜度 | 說明 |
| :--- | :--- | :--- | :--- | :--- |
| **新增並配對** | `POST /api/v1/people/match` | `handleAddSinglePersonAndMatch` | O(S_opp) | 包含 O(1) 唯一性檢查與異性池一趟掃描 (Single Pass)。 |
| **移除人員** | `DELETE /api/v1/people/{name}` | `handleRemoveSinglePerson` | O(1) | 分區 Map 雜湊檢索，常數時間定位。 |
| **查詢配對名單** | `GET /api/v1/people/{name}/matches` | `handleQueryPersonMatches` | O(S_opp + K log K) | 異性池篩選 O(S_opp) 與結果排序 O(K log K)。 |
| **查詢單一成員** | `GET /api/v1/people/{name}` | `handleQuerySinglePerson` | O(1) | 透過名稱雜湊直接檢索。 |
| **查詢所有成員** | `GET /api/v1/people` | `handleQuerySinglePeople` | O(S log S) | 全量資料提取 O(S) 並進行穩定排序。 |

## 5. 待優化事項 (TBD)
- **唯一識別碼優化**：預計改用 **UUID/ULID** 作為內部 Primary Key，解決 `name` 同名衝突問題。
- **搜尋索引優化**：計畫建立 **配對權重機制**。
- **分頁處理**：為查詢 API 增加 `limit/offset` 支援，避免大數據量下單次傳輸 Payload 過大。
