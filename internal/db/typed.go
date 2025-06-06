// Copyright (C) 2014 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package db

import (
	"database/sql"
	"encoding/binary"
	"errors"
	"time"
)

// Typed is a simple key-value store using a specific namespace within a
// lower level KV.
type Typed struct {
	db     KV
	prefix string
}

func NewMiscDB(db KV) *Typed {
	return NewTyped(db, "misc")
}

// NewTyped returns a new typed key-value store that lives in the namespace
// specified by the prefix.
func NewTyped(db KV, prefix string) *Typed {
	return &Typed{
		db:     db,
		prefix: prefix,
	}
}

// PutInt64 stores a new int64. Any existing value (even if of another type)
// is overwritten.
func (n *Typed) PutInt64(key string, val int64) error {
	var valBs [8]byte
	binary.BigEndian.PutUint64(valBs[:], uint64(val)) //nolint:gosec
	return n.db.PutKV(n.prefixedKey(key), valBs[:])
}

// Int64 returns the stored value interpreted as an int64 and a boolean that
// is false if no value was stored at the key.
func (n *Typed) Int64(key string) (int64, bool, error) {
	valBs, err := n.db.GetKV(n.prefixedKey(key))
	if err != nil {
		return 0, false, filterNotFound(err)
	}
	val := binary.BigEndian.Uint64(valBs)
	return int64(val), true, nil //nolint:gosec
}

// PutTime stores a new time.Time. Any existing value (even if of another
// type) is overwritten.
func (n *Typed) PutTime(key string, val time.Time) error {
	valBs, _ := val.MarshalBinary() // never returns an error
	return n.db.PutKV(n.prefixedKey(key), valBs)
}

// Time returns the stored value interpreted as a time.Time and a boolean
// that is false if no value was stored at the key.
func (n *Typed) Time(key string) (time.Time, bool, error) {
	var t time.Time
	valBs, err := n.db.GetKV(n.prefixedKey(key))
	if err != nil {
		return t, false, filterNotFound(err)
	}
	err = t.UnmarshalBinary(valBs)
	return t, err == nil, err
}

// PutString stores a new string. Any existing value (even if of another type)
// is overwritten.
func (n *Typed) PutString(key, val string) error {
	return n.db.PutKV(n.prefixedKey(key), []byte(val))
}

// String returns the stored value interpreted as a string and a boolean that
// is false if no value was stored at the key.
func (n *Typed) String(key string) (string, bool, error) {
	valBs, err := n.db.GetKV(n.prefixedKey(key))
	if err != nil {
		return "", false, filterNotFound(err)
	}
	return string(valBs), true, nil
}

// PutBytes stores a new byte slice. Any existing value (even if of another type)
// is overwritten.
func (n *Typed) PutBytes(key string, val []byte) error {
	return n.db.PutKV(n.prefixedKey(key), val)
}

// Bytes returns the stored value as a raw byte slice and a boolean that
// is false if no value was stored at the key.
func (n *Typed) Bytes(key string) ([]byte, bool, error) {
	valBs, err := n.db.GetKV(n.prefixedKey(key))
	if err != nil {
		return nil, false, filterNotFound(err)
	}
	return valBs, true, nil
}

// PutBool stores a new boolean. Any existing value (even if of another type)
// is overwritten.
func (n *Typed) PutBool(key string, val bool) error {
	if val {
		return n.db.PutKV(n.prefixedKey(key), []byte{0x0})
	}
	return n.db.PutKV(n.prefixedKey(key), []byte{0x1})
}

// Bool returns the stored value as a boolean and a boolean that
// is false if no value was stored at the key.
func (n *Typed) Bool(key string) (bool, bool, error) {
	valBs, err := n.db.GetKV(n.prefixedKey(key))
	if err != nil {
		return false, false, filterNotFound(err)
	}
	return valBs[0] == 0x0, true, nil
}

// Delete deletes the specified key. It is allowed to delete a nonexistent
// key.
func (n *Typed) Delete(key string) error {
	return n.db.DeleteKV(n.prefixedKey(key))
}

func (n *Typed) prefixedKey(key string) string {
	return n.prefix + "/" + key
}

func filterNotFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return err
}
