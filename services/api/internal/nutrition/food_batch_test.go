package nutrition

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/store"
)

func TestBatchFoodAllOrNothingAndRetries(t *testing.T) {
	weight:=80.0
	svc,_,userID:=nutritionFixture(t,&weight)
	setAutoNutrition(t,svc,userID)
	ctx:=context.Background()
	when:=time.Date(2026,9,20,12,0,0,0,time.UTC)
	key:="food-test-batch-01"
	invalid:=[]FoodBatchItem{{FoodID:"banana",QuantityG:120},{FoodID:"no-such-food",QuantityG:80}}
	if _,err:=svc.LogFoodBatch(ctx,userID,"breakfast",invalid,&when,key,"ai_confirm","");!errors.Is(err,store.ErrNotFound){t.Fatalf("missing food: %v",err)}
	day,err:=svc.Day(ctx,userID,when)
	if err!=nil || len(day.Entries)!=0 {t.Fatalf("partial batch saved: entries=%d err=%v",len(day.Entries),err)}
	items:=[]FoodBatchItem{{FoodID:"banana",QuantityG:120},{FoodID:"oats-dry",QuantityG:80}}
	first,err:=svc.LogFoodBatch(ctx,userID,"breakfast",items,&when,key,"ai_confirm","")
	if err!=nil || len(first.Entries)!=2 {t.Fatalf("first batch=%d err=%v",len(first.Entries),err)}
	retry,err:=svc.LogFoodBatch(ctx,userID,"breakfast",items,&when,key,"ai_confirm","")
	if err!=nil || len(retry.Entries)!=2 {t.Fatalf("same-key retry duplicated entries=%d err=%v",len(retry.Entries),err)}
	if _,err:=svc.LogFoodBatch(ctx,userID,"snack",items,&when,key,"ai_confirm","");!errors.Is(err,store.ErrIdempotencyConflict){t.Fatalf("changed meal should conflict: %v",err)}
	second,err:=svc.LogFoodBatch(ctx,userID,"breakfast",items,&when,"food-test-batch-02","ai_confirm","")
	if err!=nil || len(second.Entries)!=4 {t.Fatalf("new action must log again entries=%d err=%v",len(second.Entries),err)}
	if _,err:=svc.LogFoodBatch(ctx,userID,"breakfast",items,&when,"short","recipe:other","");err==nil {t.Fatal("expected invalid key rejection")}
}

func TestRecipeLoggingIsAtomicAndKeyed(t *testing.T) {
	weight:=80.0
	svc,st,userID:=nutritionFixture(t,&weight)
	setAutoNutrition(t,svc,userID)
	ctx:=context.Background()
	when:=time.Date(2026,9,20,12,0,0,0,time.UTC)
	corrupt,err:=st.CreateRecipe(ctx,store.Recipe{UserID:userID,Name:"Broken",Items:[]store.RecipeItem{{FoodID:"banana",QuantityG:120},{FoodID:"no-such-food",QuantityG:50}}})
	if err!=nil {t.Fatal(err)}
	if _,err:=svc.LogRecipe(ctx,userID,corrupt.ID,"lunch",1,&when,"recipe-failed-01");!errors.Is(err,store.ErrNotFound){t.Fatalf("invalid recipe: %v",err)}
	day,err:=svc.Day(ctx,userID,when)
	if err!=nil || len(day.Entries)!=0 {t.Fatalf("partial recipe saved entries=%d err=%v",len(day.Entries),err)}
	recipe,err:=svc.CreateRecipe(ctx,userID,RecipeInput{Name:"Breakfast",Items:[]RecipeItemInput{{FoodID:"banana",QuantityG:120},{FoodID:"oats-dry",QuantityG:80}}})
	if err!=nil {t.Fatal(err)}
	for i:=0;i<2;i++ {
		day,err=svc.LogRecipe(ctx,userID,recipe.ID,"lunch",1,&when,"recipe-action-01")
		if err!=nil || len(day.Entries)!=2 {t.Fatalf("recipe retry %d entries=%d err=%v",i,len(day.Entries),err)}
	}
	day,err=svc.LogRecipe(ctx,userID,recipe.ID,"lunch",1,&when,"recipe-action-02")
	if err!=nil || len(day.Entries)!=4 {t.Fatalf("second user action entries=%d err=%v",len(day.Entries),err)}
}
