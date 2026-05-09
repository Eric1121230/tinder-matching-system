package gateway

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/tinder/internal/model"
	"example.com/tinder/internal/repository"
	"example.com/tinder/internal/service"
)

func setup() *HTTPGateway {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	repo := repository.NewInMemoryPersonRepository()
	svc := service.NewMatchingService(repo)
	return NewHTTPGateway(logger, svc)
}

func TestGateway_handleAddSinglePersonAndMatch(t *testing.T) {
	gw := setup()
	server := gw.Handler()

	// 1. 測試：請求格式錯誤 (INVALID_JSON -> 400)
	t.Run("INVALID_JSON_400", func(t *testing.T) {
		// 故意傳入壞掉的 JSON 格式
		badBody := `{"name": "Female_1", "height": 160,`
		req := httptest.NewRequest("POST", "/api/v1/people/match", strings.NewReader(badBody))
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("預期 400, 得到 %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "INVALID_JSON") {
			t.Errorf("應回傳 INVALID_JSON 錯誤代碼, 得到 %s", w.Body.String())
		}
	})

	// 2. 測試：唯一性檢查 (USER_ALREADY_EXISTS -> 409)
	t.Run("USER_ALREADY_EXISTS_409", func(t *testing.T) {
		body := `{"name":"Female_1","height":160,"gender":"female","wanted_dates":2}`

		// 第一次新增成功
		req1 := httptest.NewRequest("POST", "/api/v1/people/match", strings.NewReader(body))
		w1 := httptest.NewRecorder()
		server.ServeHTTP(w1, req1)

		if w1.Code != http.StatusOK {
			t.Errorf("應該為 200成功, 得到 %d", w1.Code)
		}
		// 第二次新增同名 Female_1
		req2 := httptest.NewRequest("POST", "/api/v1/people/match", strings.NewReader(body))
		w2 := httptest.NewRecorder()
		server.ServeHTTP(w2, req2)

		if w2.Code != http.StatusConflict {
			t.Errorf("重複姓名應回傳 409, 得到 %d", w2.Code)
		}
		if !strings.Contains(w2.Body.String(), "USER_ALREADY_EXISTS") {
			t.Errorf("應回傳 USER_ALREADY_EXISTS 錯誤代碼, 得到 %s", w2.Body.String())
		}
	})

	// 3. 測試：驗證失敗 (ADD_PERSON_FAILED -> 400)
	t.Run("ADD_PERSON_FAILED_400", func(t *testing.T) {
		// 假設 wanted_dates 為 0 會觸發 service 的驗證錯誤
		invalidBody := `{"name":"Male_1","height":180,"gender":"male","wanted_dates":0}`
		req := httptest.NewRequest("POST", "/api/v1/people/match", strings.NewReader(invalidBody))
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("驗證失敗應回傳 400, 得到 %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "ADD_PERSON_FAILED") {
			t.Errorf("應回傳 ADD_PERSON_FAILED 錯誤代碼, 得到 %s", w.Body.String())
		}
	})
}
func TestGateway_handleRemoveSinglePerson(t *testing.T) {
	gw := setup()
	server := gw.Handler()

	// 先準備資料：新增 Female_1 以供後續刪除測試
	seedBody := `{"name":"Female_1","height":160,"gender":"female","wanted_dates":2}`
	reqSeed := httptest.NewRequest("POST", "/api/v1/people/match", strings.NewReader(seedBody))
	server.ServeHTTP(httptest.NewRecorder(), reqSeed)

	// 1. 測試：成功刪除使用者 (Success -> 200)
	t.Run("REMOVE_SUCCESS_200", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/v1/people/Female_1", nil)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
			t.Errorf("預期成功 200, 得到 %d", w.Code)
		}
	})

	// 2. 測試：刪除不存在的人 (PERSON_NOT_FOUND -> 404)
	t.Run("PERSON_NOT_FOUND_404", func(t *testing.T) {
		// 刪除一個不存在的名字 Ghost
		req := httptest.NewRequest("DELETE", "/api/v1/people/Ghost", nil)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("找不到人應回傳 404, 得到 %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "PERSON_NOT_FOUND") {
			t.Errorf("應回傳 PERSON_NOT_FOUND 錯誤代碼, 得到 %s", w.Body.String())
		}
	})

	// 3. 測試：姓名參數為空或無效 (INVALID_PARAMS -> 400)
	t.Run("EMPTY_NAME_400", func(t *testing.T) {
		// 測試路徑參數缺失或僅為空格的情況（取決於你的路由器實現）
		req := httptest.NewRequest("DELETE", "/api/v1/people/", nil)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		// 這裡通常會回傳 400 或 404
		if w.Code == http.StatusOK {
			t.Errorf("無效姓名不應回傳 200")
		}
	})
}

func TestGateway_handleQuerySinglePerson(t *testing.T) {
	gw := setup()
	server := gw.Handler()

	// 先準備資料
	seedBody := `{"name":"Feamale_1","height":160,"gender":"female","wanted_dates":2}`
	reqSeed := httptest.NewRequest("POST", "/api/v1/people/match", strings.NewReader(seedBody))
	server.ServeHTTP(httptest.NewRecorder(), reqSeed)

	// 1. 測試：成功查詢單人 (SUCCESS -> 200)
	t.Run("QUERY_SUCCESS_200", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/people/Feamale_1", nil)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("預期 200, 得到 %d", w.Code)
		}

		// 驗證回傳的 JSON 包含正確姓名
		if !strings.Contains(w.Body.String(), `"name":"Feamale_1"`) {
			t.Errorf("回傳內容不正確: %s", w.Body.String())
		}
	})

	// 2. 測試：找不到該人員 (PERSON_NOT_FOUND -> 404)
	t.Run("PERSON_NOT_FOUND_404", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/people/NoOne", nil)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("預期 404, 得到 %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "PERSON_NOT_FOUND") {
			t.Errorf("應包含錯誤代碼 PERSON_NOT_FOUND, 得到 %s", w.Body.String())
		}
	})

	// 3. 測試：姓名為空或僅有空格 (MISSING_NAME -> 400)
	t.Run("MISSING_NAME_400", func(t *testing.T) {
		// %20 代表空格，會被 TrimSpace 刷掉變成 ""
		req := httptest.NewRequest("GET", "/api/v1/people/%20", nil)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("預期 400, 得到 %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "MISSING_NAME") {
			t.Errorf("應包含錯誤代碼 MISSING_NAME, 得到 %s", w.Body.String())
		}
	})
}
func TestGateway_handleQuerySinglePeople(t *testing.T) {
	gw := setup()
	server := gw.Handler()

	// 1. 測試：當池子是空的時候 (SUCCESS -> 200)
	t.Run("QUERY_EMPTY_200", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/people", nil)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("預期 200, 得到 %d", w.Code)
		}
		// 應回傳封裝過的空列表 {"people":[]}
		if !strings.Contains(w.Body.String(), `"people":[]`) {
			t.Errorf("空池子應回傳空列表格式, 得到 %s", w.Body.String())
		}
	})

	// 2. 測試：有多筆資料時的封裝與排序 (SUCCESS -> 200)
	t.Run("QUERY_LIST_SORTED_200", func(t *testing.T) {
		// 準備資料：故意亂序新增
		people := []string{
			`{"name":"Female_1","height":170,"gender":"female","wanted_dates":2}`,
			`{"name":"Female_2","height":160,"gender":"female","wanted_dates":2}`,
			`{"name":"Male_1","height":150,"gender":"male","wanted_dates":2}`,
		}
		for _, p := range people {
			reqSeed := httptest.NewRequest("POST", "/api/v1/people/match", strings.NewReader(p))
			server.ServeHTTP(httptest.NewRecorder(), reqSeed)
		}

		req := httptest.NewRequest("GET", "/api/v1/people", nil)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("預期 200, 得到 %d", w.Code)
		}

		// 驗證回傳內容
		var resp model.QuerySinglePeopleResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("無法解析回傳的 JSON: %v", err)
		}

		if len(resp.People) != 3 {
			t.Errorf("預期 3 人, 得到 %d 人", len(resp.People))
		}

		if resp.People[0].Name != "Female_1" || resp.People[2].Name != "Male_1" {
			t.Errorf("排序不正確, 第一位應為 Female_1, 最後一位應為 Male_1")
		}
	})
}
func TestGateway_handleQueryPersonMatches(t *testing.T) {
	gw := setup()
	server := gw.Handler()

	people := []string{
		`{"name":"Feamale_1","height":160,"gender":"female","wanted_dates":2}`,
		`{"name":"Male_1","height":180,"gender":"male","wanted_dates":2}`,
	}
	for _, p := range people {
		server.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("POST", "/api/v1/people/match", strings.NewReader(p)))
	}

	// 1. 成功查詢潛在配對 (SUCCESS -> 200)
	t.Run("QUERY_MATCHES_SUCCESS_200", func(t *testing.T) {
		// 查詢 Male_1 的配對，預期會看到 Feamale_1
		req := httptest.NewRequest("GET", "/api/v1/people/Male_1/matches?top=5", nil)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("預期 200, 得到 %d", w.Code)
		}

		var resp model.QueryPersonMatchesResponse
		json.Unmarshal(w.Body.Bytes(), &resp)

		if resp.Name != "Male_1" || len(resp.Matches) == 0 {
			t.Error("應回傳 Male_1 的配對資料且 Matches 不應為空")
		}
	})

	// 2. 測試：缺少 top 參數 (MISSING_TOP -> 400)
	t.Run("MISSING_TOP_400", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/people/Male_1/matches", nil) // 故意不帶 ?top=
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("缺少 top 應回傳 400, 得到 %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "MISSING_TOP") {
			t.Errorf("應包含錯誤代碼 MISSING_TOP, 得到 %s", w.Body.String())
		}
	})

	// 3. 測試：top 參數非法 (INVALID_TOP -> 400)
	t.Run("INVALID_TOP_400", func(t *testing.T) {
		invalidTops := []string{"abc", "0", "-5"}
		for _, val := range invalidTops {
			url := "/api/v1/people/Male_1/matches?top=" + val
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			server.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("top=%s 應回傳 400, 得到 %d", val, w.Code)
			}
		}
	})

	// 4. 測試：找不到該人員 (PERSON_NOT_FOUND -> 404)
	t.Run("PERSON_NOT_FOUND_404", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/people/Ghost/matches?top=1", nil)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("找不到人應回傳 404, 得到 %d", w.Code)
		}
	})
}
