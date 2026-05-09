package service

import (
	"sync"
	"testing"

	"example.com/tinder/internal/model"
	"example.com/tinder/internal/repository"
)

// 基礎商業邏輯測試
// 1.AddSinglePersonAndMatch後  配對次數2次
// 2.再一次AddSinglePersonAndMatch 對1而言異性 但不符合條件 配對次數1次
// 3.再一次 AddSinglePersonAndMatch  對1而言異性但符合條件 配對次數3次
// 4.查詢1的最優配對人選
// 5.再一次 AddSinglePersonAndMatch  對1而言異性但符合條件 配對次數3次  /再做一次4
// 6.刪除剩下隨便一個人再 查詢刪除的人/全部人
func TestMatchingService_FunctionalScenario(t *testing.T) {
	repo := repository.NewInMemoryPersonRepository()
	svc := NewMatchingService(repo)

	// 1. Add Female_1 (Female, 160cm), WantedDates: 2
	// 預期：池子有 Female_1，次數為 2
	t.Run("Scenario_1_Add_Female_1", func(t *testing.T) {
		Female_1 := model.Person{Name: "Female_1", Height: 160, Gender: model.GenderFemale, WantedDates: 2}
		svc.AddSinglePersonAndMatch(Female_1)

		p, ok := svc.QuerySinglePerson("Female_1")
		if !ok || p.RemainingDates != 2 {
			t.Errorf("Female_1 should be in repo with 2 dates")
		}
		if len(svc.QuerySinglePeople()) != 1 {
			t.Error("Total people should be 1")
		}
	})

	// 2. Add ShortBoy (Male, 150cm), WantedDates: 1
	// 預期：與 Female_1 不匹配 (男比女矮)，池子增加 ShortBoy，Female_1 次數仍為 2
	t.Run("Scenario_2_Add_Incompatible_Male", func(t *testing.T) {
		shortBoy := model.Person{Name: "ShortBoy", Height: 150, Gender: model.GenderMale, WantedDates: 1}
		matches, _ := svc.AddSinglePersonAndMatch(shortBoy)

		if len(matches) != 0 {
			t.Error("Should not match ShortBoy and Female_1")
		}
		p, ok := svc.QuerySinglePerson("Female_1")
		if !ok {
			t.Error("Female_1 Not Found")
		}
		if p.RemainingDates != 2 {
			t.Error("Female_1 dates should still be 2")
		}
		if len(svc.QuerySinglePeople()) != 2 {
			t.Error("Total people should be 2")
		}
	})

	// 3. Add Male_1 (Male, 180cm), WantedDates: 3
	// 預期：符合 Female_1 條件，配對成功。Female_1 剩 1 次，Male_1 剩 2 次 (3-1=2)
	t.Run("Scenario_3_Add_Compatible_Male_Male_1", func(t *testing.T) {
		Male_1 := model.Person{Name: "Male_1", Height: 180, Gender: model.GenderMale, WantedDates: 3}
		matches, _ := svc.AddSinglePersonAndMatch(Male_1)

		if len(matches) != 1 || matches[0].GirlName != "Female_1" {
			t.Error("Male_1 should match Female_1")
		}
		a, _ := svc.QuerySinglePerson("Female_1")
		b, _ := svc.QuerySinglePerson("Male_1")
		if a.RemainingDates != 1 || b.RemainingDates != 2 {
			t.Errorf("Dates mismatch: Female_1(%d), Male_1(%d)", a.RemainingDates, b.RemainingDates)
		}
	})

	// 4. Query Female_1's potential matches
	// 預期：潛在對象應該包含 Male_1 (因為 Male_1 比 Female_1 高且是男性)
	t.Run("Scenario_4_Query_Female_1_Potential_Matches", func(t *testing.T) {
		matches, ok := svc.QueryPersonMatches("Female_1", 10)
		if !ok || len(matches) != 1 || matches[0].Name != "Male_1" {
			t.Errorf("Female_1's potential match should be Male_1, got: %+v", matches)
		}
	})

	// 5. Add Male_2 (Male, 175cm), WantedDates: 3
	// 預期：與 Female_1 配對。Female_1 次數歸 0 (1-1=0) 並消失。Male_2 剩 2 次 (3-1=2)
	t.Run("Scenario_5_Add_Compatible_Male_Male_2", func(t *testing.T) {
		Male_2 := model.Person{Name: "Male_2", Height: 175, Gender: model.GenderMale, WantedDates: 3}
		matches, _ := svc.AddSinglePersonAndMatch(Male_2)

		if len(matches) != 1 || matches[0].GirlName != "Female_1" {
			t.Error("Male_2 should match Female_1")
		}

		// 檢查 Female_1 是否消失
		if _, ok := svc.QuerySinglePerson("Female_1"); ok {
			t.Error("Female_1 should be removed as dates reached 0")
		}

		// 再次查詢 Male_2 的潛在配對 (此時 Female_1 已消失，池子剩 ShortBoy 和 Male_1)
		// Male_2(男) 的配對清單應該為空 (池子沒女性了)
		cMatches, ok := svc.QueryPersonMatches("Male_2", 10)
		if !ok {
			t.Fatal("Male_2 應該要在池子裡")
		}
		if len(cMatches) != 0 {
			t.Errorf("池子裡已經沒有女性了，但 Male_2 卻抓到 %d 個配對", len(cMatches))
		}

		allPeople := repo.ListAll()
		t.Logf("當前總人數: %d (預期應為 3: ShortBoy, Male_1, Male_2)", len(allPeople))
		if len(allPeople) != 3 {
			t.Errorf("總人數不符，預期 3 人，實際 %d 人", len(allPeople))
		}

		males := repo.ListByGender(model.GenderMale)
		if len(males) != 3 {
			t.Errorf("男性人數不符，預期 3 人，實際 %d 人", len(allPeople))
		}
		female := repo.ListByGender(model.GenderFemale)
		if len(female) != 0 {
			t.Errorf("女性人數不符，預期 0 人，實際 %d 人", len(allPeople))
		}
	})

	// 6. Delete ShortBoy
	// 預期：ShortBoy 被移除，池子只剩 Male_1 和 Male_2
	t.Run("Scenario_6_Delete_ShortBoy", func(t *testing.T) {
		if !svc.RemoveSinglePerson("ShortBoy") {
			t.Error("Failed to remove ShortBoy")
		}
		if _, ok := svc.QuerySinglePerson("ShortBoy"); ok {
			t.Error("ShortBoy should be gone")
		}
		if len(svc.QuerySinglePeople()) != 2 {
			t.Errorf("Total people should be 2 (Male_1, Male_2), got %d", len(svc.QuerySinglePeople()))
		}
	})
}

// 基礎商業邏輯測試-錯誤
// 1.AddSinglePersonAndMatch 結構失敗
// 2.AddSinglePersonAndMatch 重複名稱
// 3.尋找 / 查詢配對TOP / 刪除 不存在的名稱
// 4.測試 Top 參數為負數或零的邊界情況
// 5.測試 新增後馬上就配成功 是否還會入庫
// 6.高併發狀態下原子性扣除是否正常
func TestMatchingService_FunctionalErrors(t *testing.T) {
	repo := repository.NewInMemoryPersonRepository()
	svc := NewMatchingService(repo)

	// 1. 測試資料驗證失敗 (Validate error)
	t.Run("Error_1_Invalid_Data", func(t *testing.T) {
		// 假設你的 model.Validate 會檢查名字不能為空
		invalidP := model.Person{Name: "", Height: 170}
		_, err := svc.AddSinglePersonAndMatch(invalidP)
		if err == nil {
			t.Error("預期無效資料應報錯，但卻成功了")
		}

		WantedDates_0 := model.Person{Name: "WantedDates_0", Height: 160, Gender: model.GenderFemale, WantedDates: 0}
		_, err = svc.AddSinglePersonAndMatch(WantedDates_0)
		if err == nil {
			t.Error("預期無效資料應報錯，但卻成功了")
		}

		Height_0 := model.Person{Name: "Height_0", Height: -1, Gender: model.GenderFemale, WantedDates: 1}
		_, err = svc.AddSinglePersonAndMatch(Height_0)
		if err == nil {
			t.Error("預期無效資料應報錯，但卻成功了")
		}

		Gender := model.Person{Name: "Gender", Height: 160, Gender: "test", WantedDates: 1}
		_, err = svc.AddSinglePersonAndMatch(Gender)
		if err == nil {
			t.Error("預期無效資料應報錯，但卻成功了")
		}
	})

	// 2. 測試重複新增 (user already exists)
	t.Run("Error_2_Duplicate_User_Should_Not_Update", func(t *testing.T) {
		repo := repository.NewInMemoryPersonRepository()
		svc := NewMatchingService(repo)

		// 1. 先加入一個 Male_1，次數為 1
		p := model.Person{Name: "Male_1", Height: 170, Gender: model.GenderMale, WantedDates: 1}
		svc.AddSinglePersonAndMatch(p)

		// 2. 企圖用同名、但次數改為 3 的資料再次新增
		p.WantedDates = 3
		_, err := svc.AddSinglePersonAndMatch(p)

		// 預期應報錯
		if err == nil {
			t.Error("預期重複新增應報錯，但卻成功了")
		}

		// 3. 核心驗證：檢查原本池子裡的 Male_1 狀態
		if p1, ok := svc.QuerySinglePerson("Male_1"); ok {
			// 這裡要用 RemainingDates，且預期應該維持 1
			if p1.RemainingDates == 3 {
				t.Error("安全性漏洞：重複新增的請求不應修改原始資料的剩餘次數")
			}
			if p1.RemainingDates != 1 {
				t.Errorf("預期次數應維持 1，實際為 %d", p1.RemainingDates)
			}
		} else {
			t.Error("原本的人不見了，這不符合預期")
		}
	})

	// 3. 測試查詢不存在的使用者
	t.Run("Error_3_Query_Non_Existent", func(t *testing.T) {
		if _, ok := svc.QuerySinglePerson("Ghost"); ok {
			t.Error("預期找不到人，但 ok 為 true")
		}

		if _, ok := svc.QueryPersonMatches("Ghost", 10); ok {
			t.Error("預期找不到配對來源，但 ok 為 true")
		}
		if svc.RemoveSinglePerson("Ghost") {
			t.Error("預期刪除不存在的人應回傳 false")
		}
	})

	// 4. 測試 Top 參數為負數或零的邊界情況
	t.Run("Error_4_Top_Boundary", func(t *testing.T) {
		svc.AddSinglePersonAndMatch(model.Person{Name: "Male_1", Gender: model.GenderMale, Height: 180, WantedDates: 5})

		res, ok := svc.QueryPersonMatches("Male_1", 0)
		if !ok || len(res) != 0 {
			t.Error("Top 為 0 時應回傳空陣列且 ok")
		}

		resNeg, _ := svc.QueryPersonMatches("Male_1", -5)
		if len(resNeg) != 0 {
			t.Error("Top 為負數時應回傳空陣列")
		}
	})
	// 5. 測試 新增後馬上就配成功
	t.Run("Error_5_Immediate_Match_No_Storage", func(t *testing.T) {
		svc.AddSinglePersonAndMatch(model.Person{Name: "Male_1", Gender: model.GenderMale, Height: 180, WantedDates: 5})
		svc.AddSinglePersonAndMatch(model.Person{Name: "Male_2", Gender: model.GenderMale, Height: 180, WantedDates: 5})
		svc.AddSinglePersonAndMatch(model.Person{Name: "Female_1", Gender: model.GenderFemale, Height: 170, WantedDates: 1})

		allPeople := repo.ListAll()
		t.Logf("當前總人數: %d (預期應為 2:  Male_1, Male_2)", len(allPeople))
		if len(allPeople) != 2 {
			t.Errorf("總人數不符，預期 2 人，實際 %d 人", len(allPeople))
		}

		males := repo.ListByGender(model.GenderMale)
		if len(males) != 2 {
			t.Errorf("男性人數不符，預期 2 人，實際 %d 人", len(allPeople))
		}
		var totalDates int
		for _, m := range males {
			totalDates += m.WantedDates
		}
		if totalDates != 9 {
			t.Errorf("總次數不符，預期 9 次，實際 %d 次", totalDates)
		}
		female := repo.ListByGender(model.GenderFemale)
		if len(female) != 0 {
			t.Errorf("女性人數不符，預期 0 人，實際 %d 人", len(female))
		}
	})
	t.Run("Scenario_6_Precise_Concurrency_Check", func(t *testing.T) {
		repo := repository.NewInMemoryPersonRepository()
		svc := NewMatchingService(repo)

		svc.AddSinglePersonAndMatch(model.Person{Name: "Female_1", Gender: model.GenderFemale, Height: 160, WantedDates: 1})

		winnerChan := make(chan string, 3)
		var wg sync.WaitGroup

		contestants := []model.Person{
			{Name: "Male_1", Gender: model.GenderMale, Height: 180, WantedDates: 1},
			{Name: "Male_2", Gender: model.GenderMale, Height: 180, WantedDates: 1},
			{Name: "Male_3", Gender: model.GenderMale, Height: 180, WantedDates: 2},
		}

		wg.Add(len(contestants))
		for _, p := range contestants {
			go func(person model.Person) {
				defer wg.Done()
				// 增加對 err 的檢查
				matches, err := svc.AddSinglePersonAndMatch(person)
				if err != nil {
					t.Errorf("併發請求失敗: %v", err)
					return
				}
				if len(matches) > 0 {
					winnerChan <- person.Name
				}
			}(p)
		}

		wg.Wait()
		close(winnerChan)

		// 安全地取得贏家
		winnerName, ok := <-winnerChan
		if !ok {
			t.Fatal("沒有任何男生成功配對到 Female_1")
		}

		males := repo.ListByGender(model.GenderMale)

		// 1. 驗證剩餘人數
		// Male_3 贏: 2-1=1 (留任) + 其他 2 人 = 3 人
		// Male_1/2 贏: 1-1=0 (消失) + 其他 2 人 = 2 人
		expectedCount := 2
		if winnerName == "Male_3" {
			expectedCount = 3
		}

		if len(males) != expectedCount {
			t.Errorf("贏家是 %s, 預期 Repo 剩 %d 人, 實際得到 %d 人", winnerName, expectedCount, len(males))
		}

		// 2. 驗證總次數消耗 (不變量 Invariant)
		// 初始: 1+1+2=4. 配對消耗 1 次後必為 3
		var totalDates int
		for _, m := range males {
			totalDates += m.WantedDates
		}
		if totalDates != 3 {
			t.Errorf("總剩餘次數對帳失敗: 預期 3, 實際 %d", totalDates)
		}
	})

}
