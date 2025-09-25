package hashcache

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestNew(t *testing.T) {

	type args struct {
		putKey interface{}
		getKey interface{}
		value  interface{}
	}
	type test struct {
		name       string
		args       args
		wantPuErr  bool
		want       interface{}
		wantGetErr bool
	}
	type keyStruct1 struct {
		a int
	}
	type keyStruct2 struct {
		A int
		B float64
		C interface{}
		D string
	}
	type keyStruct3 struct {
		A int
		B *keyStruct2
	}

	tests := []test{
		{
			name: "Int key",
			args: args{
				putKey: 123,
				getKey: 123,
				value:  "Hello world",
			},
			want: "Hello world",
		},
		{
			name: "[]byte key",
			args: args{
				putKey: []byte{'a', 'k', 'e', 'y'},
				getKey: []byte{'a', 'k', 'e', 'y'},
				value:  "Hello world2",
			},
			want: "Hello world2",
		},
		{
			name: "float key",
			args: args{
				putKey: 1.234,
				getKey: 1.234,
				value:  "Hello world3",
			},
			want: "Hello world3",
		},
		{
			name: "struct key - no exported attributes should error",
			args: args{
				putKey: keyStruct1{a: 2},
				value:  "Hello world4",
			},
			wantPuErr: true,
		},
		{
			name: "struct key",
			args: args{
				putKey: keyStruct2{A: 2, B: 3.4, C: nil, D: "Hello"},
				getKey: keyStruct2{A: 2, B: 3.4, C: nil, D: "Hello"},
				value:  "Hello world5",
			},
			want: "Hello world5",
		},
		{
			name: "struct key containing references",
			args: args{
				putKey: keyStruct3{A: 5, B: &keyStruct2{A: 2, B: 3.4, C: nil, D: "Hello"}},
				getKey: keyStruct3{A: 5, B: &keyStruct2{A: 2, B: 3.4, C: nil, D: "Hello"}},
				value:  "Hello world6",
			},
			want: "Hello world6",
		},
		{
			name: "reference to struct key containing references",
			args: args{
				putKey: &keyStruct3{A: 5, B: &keyStruct2{A: 2, B: 3.4, C: nil, D: "Hello"}},
				getKey: &keyStruct3{A: 5, B: &keyStruct2{A: 2, B: 3.4, C: nil, D: "Hello"}},
				value:  "Hello world7",
			},
			want: "Hello world7",
		},
		{
			name: "missing key",
			args: args{
				putKey: 200,
				getKey: 404,
				value:  "Hello world8",
			},
			want:       nil,
			wantGetErr: true,
		},
	}

	// Normally Cache is type specific, but can use 'any' as here to have a single cache
	cache, err := New[any](context.Background())
	if err != nil {
		t.Errorf("Unexpected error on New(): %v", err)
		return
	}
	defer cache.Delete()

	for _, tt := range tests {

		err = cache.Put(tt.args.putKey, tt.args.value)
		if (err != nil) != tt.wantPuErr {
			t.Errorf("Unexpected error on Put(): %v", err)
			continue
		}
		if tt.wantPuErr {
			continue // That was the test
		}

		resp, err := cache.Get(tt.args.getKey)
		if (err != nil) != tt.wantGetErr {
			t.Errorf("Unexpected error on Get(): %v", err)
			continue
		}

		if !reflect.DeepEqual(resp, tt.want) {
			t.Errorf("Unexpected return value from Get(): wanted %v, got %v", tt.want, resp)
		}
	}
}

func TestNew_1(t *testing.T) {

	cache, err := New[string](context.Background())
	if err != nil {
		t.Fatal(err)
	}

	cache.Delete()

	<-time.After(100 * time.Millisecond)

	err = cache.Put("AKey", "AValue")

	if err == nil {
		t.Fatal("unexpected success to Put value, after context cancelled")
	}
	if !errors.Is(err, ErrCacheInvalidated) {
		t.Fatalf("unexpected Put error: expected %v, got: %v", ErrCacheInvalidated, err)
	}

	_, err = cache.Get("AKey")

	if err == nil {
		t.Fatal("unexpected success to Get value, after context cancelled")
	}
	if !errors.Is(err, ErrCacheInvalidated) {
		t.Fatalf("unexpected error from Get: expected %v, got: %v", ErrCacheInvalidated, err)
	}
}

func TestNew_2(t *testing.T) {

	cache, err := New[string](context.Background())
	if err != nil {
		t.Fatal(err)
	}

	deleter := func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("unexpected recover: %v", r)
			}
		}()
		cache.Delete()
	}

	for range 10 {
		deleter()
	}
}

func TestNew_3(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())

	cache, err := New[string](ctx)
	if err != nil {
		t.Fatal(err)
	}

	<-time.After(100 * time.Millisecond)

	cancel()

	<-time.After(100 * time.Millisecond)

	err = cache.Put("AKey", "AValue")

	if err == nil {
		t.Fatal("unexpected success to put value, after context cancelled")
	}
	if !errors.Is(err, ErrCacheInvalidated) {
		t.Fatalf("unexpected error: expected %v, got: %v", ErrCacheInvalidated, err)
	}
}
