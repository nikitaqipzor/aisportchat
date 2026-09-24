package store

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMemoryFoodBatchTransactionAndUserScopedRetry(t *testing.T) {
	ctx:=context.Background()
	st:=NewMemory()
	first,err:=st.CreateUser(ctx,"batch-first@example.com","hash")
	if err!=nil {t.Fatal(err)}
	second,err:=st.CreateUser(ctx,"batch-second@example.com","hash")
	if err!=nil {t.Fatal(err)}
	when:=time.Date(2026,9,20,12,0,0,0,time.UTC)
	entries:=[]FoodEntry{{UserID:first.ID,FoodID:"banana",MealType:"snack",LoggedAt:when,QuantityG:100},{UserID:first.ID,FoodID:"no-such-food",MealType:"snack",LoggedAt:when,QuantityG:100}}
	if _,err:=st.CreateFoodEntries(ctx,first.ID,"shared-batch-key","hash",entries);!errors.Is(err,ErrNotFound){t.Fatalf("bad batch: %v",err)}
	all,err:=st.ListFoodEntries(ctx,first.ID,when.Add(-time.Hour),when.Add(time.Hour))
	if err!=nil || len(all)!=0 {t.Fatalf("batch was not atomic: %+v err=%v",all,err)}
	entries[1].FoodID="apple"
	if _,err:=st.CreateFoodEntries(ctx,first.ID,"shared-batch-key","hash",entries);err!=nil {t.Fatal(err)}
	if _,err:=st.CreateFoodEntries(ctx,first.ID,"shared-batch-key","hash",entries);err!=nil {t.Fatal(err)}
	// A retry may arrive hours after the original response was lost. It must
	// return the original timestamp even though a fresh attempt uses a new now.
	entries[0].LoggedAt=when.Add(25*time.Hour)
	entries[1].LoggedAt=when.Add(25*time.Hour)
	if savedAt,err:=st.CreateFoodEntries(ctx,first.ID,"shared-batch-key","hash",entries);err!=nil || !savedAt.Equal(when){t.Fatalf("late retry returned %s, err=%v",savedAt,err)}
	all,err=st.ListFoodEntries(ctx,first.ID,when.Add(-time.Hour),when.Add(time.Hour))
	if err!=nil || len(all)!=2 {t.Fatalf("duplicate entries=%d err=%v",len(all),err)}
	if _,err:=st.CreateFoodEntries(ctx,first.ID,"shared-batch-key","changed",entries);!errors.Is(err,ErrIdempotencyConflict){t.Fatalf("expected key conflict: %v",err)}
	other:=[]FoodEntry{{UserID:second.ID,FoodID:"banana",MealType:"snack",LoggedAt:when,QuantityG:100}}
	if _,err:=st.CreateFoodEntries(ctx,second.ID,"shared-batch-key","different",other);err!=nil {t.Fatalf("keys must be scoped to user: %v",err)}
}
