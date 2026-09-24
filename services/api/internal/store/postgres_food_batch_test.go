//go:build cgo

package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestPostgresFoodBatchRollsBackAndDeduplicates(t *testing.T) {
	dsn:=os.Getenv("POSTGRES_TEST_DSN")
	if dsn=="" {t.Skip("POSTGRES_TEST_DSN is not set")}
	pg,err:=NewPostgres(dsn)
	if err!=nil {t.Fatal(err)}
	defer pg.Close()
	ctx:=context.Background()
	user,err:=pg.CreateUser(ctx,fmt.Sprintf("food-batch-%d@example.com",time.Now().UnixNano()),"hash")
	if err!=nil {t.Fatal(err)}
	when:=time.Date(2026,9,20,12,0,0,0,time.UTC)
	rows:=[]FoodEntry{{UserID:user.ID,FoodID:"banana",FoodName:"Banana",MealType:"snack",LoggedAt:when,QuantityG:100,Calories:89},{UserID:user.ID,FoodID:"does-not-exist",FoodName:"Invalid",MealType:"snack",LoggedAt:when,QuantityG:100}}
	key:="postgres-food-batch-01"
	if _,err=pg.CreateFoodEntries(ctx,user.ID,key,"hash",rows);err==nil {t.Fatal("expected foreign key failure")}
	listed,err:=pg.ListFoodEntries(ctx,user.ID,when.Add(-time.Hour),when.Add(time.Hour))
	if err!=nil || len(listed)!=0 {t.Fatalf("transaction leaked entries %+v: %v",listed,err)}
	rows[1].FoodID="apple"
	rows[1].FoodName="Apple"
	for i:=0;i<2;i++ {
		got,err:=pg.CreateFoodEntries(ctx,user.ID,key,"hash",rows)
		if err!=nil || !got.Equal(when) {t.Fatalf("attempt %d time=%s err=%v",i,got,err)}
	}
	listed,err=pg.ListFoodEntries(ctx,user.ID,when.Add(-time.Hour),when.Add(time.Hour))
	if err!=nil || len(listed)!=2 {t.Fatalf("retry duplicated entries=%d err=%v",len(listed),err)}
	if _,err=pg.CreateFoodEntries(ctx,user.ID,key,"different-hash",rows);!errors.Is(err,ErrIdempotencyConflict){t.Fatalf("expected 409 conflict: %v",err)}
	if _,err=pg.CreateFoodEntries(ctx,user.ID,"postgres-food-batch-02","hash",rows);err!=nil {t.Fatal(err)}
	listed,err=pg.ListFoodEntries(ctx,user.ID,when.Add(-time.Hour),when.Add(time.Hour))
	if err!=nil || len(listed)!=4 {t.Fatalf("new action entries=%d err=%v",len(listed),err)}
	other,err:=NewPostgres(dsn)
	if err!=nil {t.Fatal(err)}
	defer other.Close()
	results:=make(chan error,2)
	go func(){_,e:=pg.CreateFoodEntries(ctx,user.ID,"postgres-concurrent-01","same-hash",rows);results<-e}()
	go func(){_,e:=other.CreateFoodEntries(ctx,user.ID,"postgres-concurrent-01","same-hash",rows);results<-e}()
	for i:=0;i<2;i++ {if e:=<-results;e!=nil {t.Fatalf("concurrent retry: %v",e)}}
	listed,err=pg.ListFoodEntries(ctx,user.ID,when.Add(-time.Hour),when.Add(time.Hour))
	if err!=nil || len(listed)!=6 {t.Fatalf("concurrent attempt duplicated entries=%d err=%v",len(listed),err)}
}

func TestPostgresConnectionLossCannotContinueTransaction(t *testing.T) {
	dsn:=os.Getenv("POSTGRES_TEST_DSN")
	if dsn=="" {t.Skip("POSTGRES_TEST_DSN is not set")}
	pg,err:=NewPostgres(dsn)
	if err!=nil {t.Fatal(err)}
	defer pg.Close()
	ctx:=context.Background()
	err=pg.withTx(ctx,func()error {
		// PostgreSQL permits a user to terminate its own backend. The next
		// statement must fail instead of silently reconnecting outside BEGIN.
		_,_=pg.queryLocked(ctx,"SELECT pg_terminate_backend(pg_backend_pid())")
		_,nextErr:=pg.queryLocked(ctx,"SELECT 1")
		return nextErr
	})
	if err==nil || !strings.Contains(err.Error(),"transaction connection was lost") {t.Fatalf("expected transaction loss without reconnect, got %v",err)}
	if _,err=pg.query(ctx,"SELECT 1");err!=nil {t.Fatalf("reconnect after failed transaction: %v",err)}
}
