package main

import (
	"context"
	"strconv"
	"testing"
)

func TestFollowingCap(t *testing.T) {
	env := newTestEnv(t)
	ctx := context.Background()
	ana := makeUser(t, env.store, "ana")
	others := make([]*User, maxFollowing+1)
	for i := range others {
		others[i] = makeUser(t, env.store, "user"+strconv.Itoa(i))
	}
	for _, o := range others[:maxFollowing] {
		if err := env.store.Follow(ctx, ana.ID, o.ID); err != nil {
			t.Fatal(err)
		}
	}
	if err := env.store.Follow(ctx, ana.ID, others[maxFollowing].ID); err != ErrFollowingTooMany {
		t.Errorf("follow #%d = %v, want ErrFollowingTooMany", maxFollowing+1, err)
	}
}
