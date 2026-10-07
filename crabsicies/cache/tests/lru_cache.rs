//! Robust LRU cache test suite (TDD red phase).
//!
//! Target API (to be implemented in `src/lru.rs`, NOT yet implemented):
//! ```rust,ignore
//! impl Cache {
//!     pub fn new(capacity: usize) -> Self;
//!     pub fn put(&mut self, key: &str, value: &str);
//!     pub fn get(&mut self, key: &str) -> Option<String>; // or Option<&String>/Option<&str>
//! }
//! ```
//! Classic LRU semantics:
//! - `get` and `put` (including overwrite of existing key) refresh recency (become MRU).
//! - `get` on missing key returns `None` and does NOT change recency.
//! - Inserting a new key when full evicts least-recently-used.
//! - Updating existing key when full does NOT evict.
//!
//! NOTE: `src/lru.rs` is intentionally left untouched per request.
//! These tests are expected to FAIL TO COMPILE until the API above exists.
//! That is normal TDD: implement `Cache` to make them pass.

// Include the implementation file directly so we don't need a lib target.
// This avoids touching src/main.rs / src/lib.rs.
#[path = "../src/lru.rs"]
mod lru;

use lru::Cache;

// ---------------------------------------------------------------------------
// Helpers (agnostic to owned vs borrowed return type of `get`)
// ---------------------------------------------------------------------------

/// Get as owned String so tests pass whether `get` returns
/// `Option<String>`, `Option<&String>` or `Option<&str>`.
fn get_owned(cache: &mut Cache, key: &str) -> Option<String> {
    cache.get(key).map(|v| v.to_string())
}

fn assert_hit(cache: &mut Cache, key: &str, expected_val: &str) {
    assert_eq!(
        get_owned(cache, key),
        Some(expected_val.to_string()),
        "expected get({key:?}) to hit with {expected_val:?}"
    );
}

fn assert_miss(cache: &mut Cache, key: &str) {
    assert_eq!(
        get_owned(cache, key),
        None,
        "expected get({key:?}) to miss (None)"
    );
}

// ---------------------------------------------------------------------------
// 1. Construction / empty behavior
// ---------------------------------------------------------------------------

#[test]
fn new_cache_get_from_empty_returns_none() {
    let mut c = Cache::new(3);
    assert_miss(&mut c, "a");
    assert_miss(&mut c, "");
    assert_miss(&mut c, "missing");
}

#[test]
fn get_missing_key_returns_none() {
    let mut c = Cache::new(2);
    c.put("a", "1");
    assert_miss(&mut c, "b");
    assert_miss(&mut c, "A"); // case sensitive
    assert_miss(&mut c, "a "); // trailing space distinct
}

#[test]
fn consecutive_misses_do_not_corrupt_cache() {
    let mut c = Cache::new(2);
    c.put("a", "1");
    assert_miss(&mut c, "x");
    assert_miss(&mut c, "y");
    assert_miss(&mut c, "z");
    // original entry still intact
    assert_hit(&mut c, "a", "1");
}

// ---------------------------------------------------------------------------
// 2. Basic put / get
// ---------------------------------------------------------------------------

#[test]
fn put_single_and_get_returns_value() {
    let mut c = Cache::new(10);
    c.put("k", "v");
    assert_hit(&mut c, "k", "v");
}

#[test]
fn put_multiple_up_to_capacity_all_retrievable() {
    let mut c = Cache::new(3);
    c.put("a", "1");
    c.put("b", "2");
    c.put("c", "3");
    assert_hit(&mut c, "a", "1");
    assert_hit(&mut c, "b", "2");
    assert_hit(&mut c, "c", "3");
}

#[test]
fn put_overwrites_existing_key_returns_new_value() {
    let mut c = Cache::new(2);
    c.put("a", "1");
    c.put("a", "2");
    assert_hit(&mut c, "a", "2");
}

#[test]
fn put_same_key_many_times_keeps_latest() {
    let mut c = Cache::new(2);
    for i in 0..50 {
        c.put("a", &format!("v{i}"));
    }
    assert_hit(&mut c, "a", "v49");
}

#[test]
fn get_after_update_returns_new_value_immediately() {
    let mut c = Cache::new(3);
    c.put("a", "1");
    c.put("b", "2");
    c.put("a", "updated");
    assert_hit(&mut c, "a", "updated");
    assert_hit(&mut c, "b", "2");
}

// ---------------------------------------------------------------------------
// 3. Capacity edge cases: 0 and 1
// ---------------------------------------------------------------------------

#[test]
fn capacity_zero_stores_nothing() {
    // Design decision under test: cap 0 holds nothing; puts are no-ops, gets always miss.
    // If your intended semantics differ (e.g. panic on 0), change this test.
    let mut c = Cache::new(0);
    c.put("a", "1");
    assert_miss(&mut c, "a");
    c.put("b", "2");
    assert_miss(&mut c, "a");
    assert_miss(&mut c, "b");
}

#[test]
fn capacity_one_holds_only_most_recent() {
    let mut c = Cache::new(1);
    c.put("a", "1");
    assert_hit(&mut c, "a", "1");
    c.put("b", "2");
    assert_miss(&mut c, "a");
    assert_hit(&mut c, "b", "2");
    c.put("c", "3");
    assert_miss(&mut c, "b");
    assert_hit(&mut c, "c", "3");
}

#[test]
fn capacity_one_update_does_not_evict() {
    let mut c = Cache::new(1);
    c.put("a", "1");
    c.put("a", "2");
    assert_hit(&mut c, "a", "2");
    // still size 1: next distinct put evicts
    c.put("b", "3");
    assert_miss(&mut c, "a");
    assert_hit(&mut c, "b", "3");
}

#[test]
fn capacity_one_get_does_not_prevent_next_eviction() {
    let mut c = Cache::new(1);
    c.put("a", "1");
    // get refreshes, but with cap 1 any new key must still evict
    assert_hit(&mut c, "a", "1");
    c.put("b", "2");
    assert_miss(&mut c, "a");
    assert_hit(&mut c, "b", "2");
}

// ---------------------------------------------------------------------------
// 4. Eviction: basic LRU
// ---------------------------------------------------------------------------

#[test]
fn evicts_lru_when_over_capacity_no_access() {
    let mut c = Cache::new(2);
    c.put("a", "1");
    c.put("b", "2");
    c.put("c", "3"); // should evict "a" (LRU)
    assert_miss(&mut c, "a");
    assert_hit(&mut c, "b", "2");
    assert_hit(&mut c, "c", "3");
}

#[test]
fn evicted_key_returns_none_and_others_survive() {
    let mut c = Cache::new(3);
    c.put("a", "1");
    c.put("b", "2");
    c.put("c", "3");
    c.put("d", "4"); // evicts a
    assert_miss(&mut c, "a");
    assert_hit(&mut c, "b", "2");
    assert_hit(&mut c, "c", "3");
    assert_hit(&mut c, "d", "4");
}

#[test]
fn many_inserts_only_last_n_remain() {
    let mut c = Cache::new(3);
    for i in 0..10 {
        c.put(&format!("k{i}"), &format!("v{i}"));
    }
    // only k7,k8,k9 should remain
    for i in 0..7 {
        assert_miss(&mut c, &format!("k{i}"));
    }
    assert_hit(&mut c, "k7", "v7");
    assert_hit(&mut c, "k8", "v8");
    assert_hit(&mut c, "k9", "v9");
}

#[test]
fn update_existing_when_full_does_not_evict() {
    let mut c = Cache::new(2);
    c.put("a", "1");
    c.put("b", "2");
    // update, not insert -> no eviction
    c.put("a", "10");
    assert_hit(&mut c, "a", "10");
    assert_hit(&mut c, "b", "2");
}

#[test]
fn many_updates_do_not_grow_or_evict() {
    let mut c = Cache::new(2);
    c.put("a", "1");
    c.put("b", "2");
    for i in 0..100 {
        c.put("a", &format!("a{i}"));
    }
    assert_hit(&mut c, "a", "a99");
    assert_hit(&mut c, "b", "2");
}

// ---------------------------------------------------------------------------
// 5. Recency: the core LRU distinction (get refreshes; FIFO would fail these)
// ---------------------------------------------------------------------------

#[test]
fn get_refreshes_recency_so_accessed_item_survives() {
    // cap 2: a,b. get(a) makes b LRU. put(c) must evict b, NOT a.
    // A FIFO cache would evict a here and fail this test.
    let mut c = Cache::new(2);
    c.put("a", "1");
    c.put("b", "2");
    assert_hit(&mut c, "a", "1"); // refresh a
    c.put("c", "3"); // evicts b
    assert_hit(&mut c, "a", "1");
    assert_miss(&mut c, "b");
    assert_hit(&mut c, "c", "3");
}

#[test]
fn get_missing_does_not_change_recency() {
    let mut c = Cache::new(2);
    c.put("a", "1");
    c.put("b", "2");
    assert_miss(&mut c, "missing"); // should not refresh anything
    c.put("c", "3"); // still evicts a (LRU)
    assert_miss(&mut c, "a");
    assert_hit(&mut c, "b", "2");
    assert_hit(&mut c, "c", "3");
}

#[test]
fn update_existing_refreshes_recency() {
    // cap 2: a,b. update a -> a is MRU, b is LRU. put c evicts b.
    let mut c = Cache::new(2);
    c.put("a", "1");
    c.put("b", "2");
    c.put("a", "10"); // refresh
    c.put("c", "3"); // evicts b
    assert_hit(&mut c, "a", "10");
    assert_miss(&mut c, "b");
    assert_hit(&mut c, "c", "3");
}

#[test]
fn lru_chain_order_three_entries() {
    // cap 3: a,b,c (LRU->MRU). get a -> b,c,a. get b -> c,a,b. put d evicts c.
    let mut c = Cache::new(3);
    c.put("a", "1");
    c.put("b", "2");
    c.put("c", "3");
    assert_hit(&mut c, "a", "1");
    assert_hit(&mut c, "b", "2");
    c.put("d", "4"); // evicts c
    assert_miss(&mut c, "c");
    assert_hit(&mut c, "a", "1");
    assert_hit(&mut c, "b", "2");
    assert_hit(&mut c, "d", "4");
}

#[test]
fn repeated_gets_keep_item_alive_across_many_evictions() {
    let mut c = Cache::new(3);
    c.put("a", "1");
    c.put("b", "2");
    c.put("c", "3");
    // keep touching a so it never becomes LRU
    assert_hit(&mut c, "a", "1");
    c.put("d", "4"); // evicts b
    assert_miss(&mut c, "b");
    assert_hit(&mut c, "a", "1"); // refresh a again
    c.put("e", "5"); // evicts c
    assert_miss(&mut c, "c");
    assert_hit(&mut c, "a", "1");
    assert_hit(&mut c, "d", "4");
    assert_hit(&mut c, "e", "5");
}

#[test]
fn access_reversal_changes_eviction_victim() {
    // cap 3: insert a,b,c. Access c,b,a so LRU->MRU becomes c,b,a.
    // Next insert d must evict c (not a).
    let mut c = Cache::new(3);
    c.put("a", "1");
    c.put("b", "2");
    c.put("c", "3");
    assert_hit(&mut c, "c", "3");
    assert_hit(&mut c, "b", "2");
    assert_hit(&mut c, "a", "1");
    c.put("d", "4"); // evicts c
    assert_miss(&mut c, "c");
    assert_hit(&mut c, "a", "1");
    assert_hit(&mut c, "b", "2");
    assert_hit(&mut c, "d", "4");
}

#[test]
fn interleaved_put_get_complex_sequence() {
    // Hand-computed LRU walkthrough, cap 3.
    let mut c = Cache::new(3);
    c.put("a", "1"); // [a]
    c.put("b", "2"); // [a,b]
    assert_hit(&mut c, "a", "1"); // [b,a]
    c.put("c", "3"); // [b,a,c]
    c.put("d", "4"); // evict b -> [a,c,d]
    assert_miss(&mut c, "b");
    assert_hit(&mut c, "c", "3"); // [a,d,c]
    c.put("a", "10"); // update a -> [d,c,a]
    c.put("e", "5"); // evict d -> [c,a,e]
    assert_miss(&mut c, "d");
    assert_miss(&mut c, "b");
    assert_hit(&mut c, "c", "3");
    assert_hit(&mut c, "a", "10");
    assert_hit(&mut c, "e", "5");
}

#[test]
fn put_after_miss_evicts_correct_lru() {
    let mut c = Cache::new(2);
    c.put("a", "1");
    c.put("b", "2");
    assert_miss(&mut c, "zzz");
    // miss changed nothing, so a is still LRU
    c.put("c", "3");
    assert_miss(&mut c, "a");
    assert_hit(&mut c, "b", "2");
    assert_hit(&mut c, "c", "3");
}

// ---------------------------------------------------------------------------
// 6. Key / value edge cases
// ---------------------------------------------------------------------------

#[test]
fn empty_string_key_and_value() {
    let mut c = Cache::new(2);
    c.put("", "");
    assert_hit(&mut c, "", "");
    c.put("", "nonempty");
    assert_hit(&mut c, "", "nonempty");
}

#[test]
fn empty_key_vs_empty_value_independent() {
    let mut c = Cache::new(3);
    c.put("", "val");
    c.put("key", "");
    assert_hit(&mut c, "", "val");
    assert_hit(&mut c, "key", "");
}

#[test]
fn unicode_keys_and_values() {
    let mut c = Cache::new(3);
    c.put("🦀", "crab");
    c.put("键", "值");
    c.put("emoji-🎉-key", "🎉🎉");
    assert_hit(&mut c, "🦀", "crab");
    assert_hit(&mut c, "键", "值");
    assert_hit(&mut c, "emoji-🎉-key", "🎉🎉");
}

#[test]
fn long_key_and_value_10k_chars() {
    let mut c = Cache::new(2);
    let long_k = "k".repeat(10_000);
    let long_v = "v".repeat(10_000);
    c.put(&long_k, &long_v);
    assert_hit(&mut c, &long_k, &long_v);
}

#[test]
fn whitespace_and_special_chars_are_distinct_keys() {
    let mut c = Cache::new(5);
    c.put("a b", "1");
    c.put("a  b", "2");
    c.put("a\nb", "3");
    c.put("a\tb", "4");
    assert_hit(&mut c, "a b", "1");
    assert_hit(&mut c, "a  b", "2");
    assert_hit(&mut c, "a\nb", "3");
    assert_hit(&mut c, "a\tb", "4");
    assert_miss(&mut c, "ab");
}

#[test]
fn keys_are_case_sensitive() {
    let mut c = Cache::new(3);
    c.put("Key", "upper");
    c.put("key", "lower");
    c.put("KEY", "allcaps");
    assert_hit(&mut c, "Key", "upper");
    assert_hit(&mut c, "key", "lower");
    assert_hit(&mut c, "KEY", "allcaps");
}

#[test]
fn similar_prefix_keys_do_not_collide() {
    let mut c = Cache::new(4);
    c.put("a", "1");
    c.put("ab", "2");
    c.put("abc", "3");
    c.put("abcd", "4");
    assert_hit(&mut c, "a", "1");
    assert_hit(&mut c, "ab", "2");
    assert_hit(&mut c, "abc", "3");
    assert_hit(&mut c, "abcd", "4");
    assert_miss(&mut c, "abcde");
    assert_miss(&mut c, "");
}

#[test]
fn numeric_string_keys() {
    let mut c = Cache::new(3);
    c.put("1", "one");
    c.put("01", "zero-one");
    c.put("1 ", "one-space");
    assert_hit(&mut c, "1", "one");
    assert_hit(&mut c, "01", "zero-one");
    assert_hit(&mut c, "1 ", "one-space");
}

// ---------------------------------------------------------------------------
// 7. Invariants, cloning, stress
// ---------------------------------------------------------------------------

#[test]
fn size_never_exceeds_capacity_via_eviction() {
    // Indirect size check using only get/put: cap 2, insert 5 distinct keys,
    // at most 2 can hit.
    let mut c = Cache::new(2);
    for i in 0..5 {
        c.put(&format!("k{i}"), &format!("v{i}"));
    }
    let hits = (0..5)
        .filter(|i| get_owned(&mut c, &format!("k{i}")).is_some())
        .count();
    assert!(
        hits <= 2,
        "cache holds {hits} items but capacity is 2 (must evict)"
    );
    // and specifically the last two survive (LRU order with no gets)
    assert_hit(&mut c, "k3", "v3");
    assert_hit(&mut c, "k4", "v4");
}

#[test]
fn returned_value_is_independent_clone() {
    let mut c = Cache::new(2);
    c.put("a", "1");
    let v1 = get_owned(&mut c, "a").expect("should hit");
    let mut mutated = v1.clone();
    mutated.push_str("-mutated");
    // mutating the returned owned value must not affect cached value
    assert_hit(&mut c, "a", "1");
    assert_ne!(mutated, get_owned(&mut c, "a").unwrap());
}

#[test]
fn overwrite_then_get_immediately_sees_new_value() {
    let mut c = Cache::new(2);
    c.put("a", "1");
    c.put("b", "2");
    for i in 0..10 {
        c.put("a", &format!("v{i}"));
        assert_hit(&mut c, "a", &format!("v{i}"));
    }
    assert_hit(&mut c, "b", "2");
}

#[test]
fn large_capacity_many_inserts() {
    let mut c = Cache::new(500);
    for i in 0..500 {
        c.put(&format!("k{i}"), &format!("v{i}"));
    }
    for i in 0..500 {
        assert_hit(&mut c, &format!("k{i}"), &format!("v{i}"));
    }
    // one more evicts k0 (LRU) — but note: the verification gets above
    // refreshed recency in order k0..k499, so LRU is k0. This still holds.
    c.put("extra", "e");
    assert_miss(&mut c, "k0");
    assert_hit(&mut c, "extra", "e");
}

#[test]
fn stress_interleaved_100_ops_ends_in_correct_state() {
    // Deterministic pseudo-random-ish sequence with hand-tracked expectation:
    // cap 4. We verify no panic, size invariant, and spot-check survivors.
    let mut c = Cache::new(4);
    for i in 0..100 {
        c.put(&format!("k{}", i % 6), &format!("v{i}"));
        // periodic reads refresh recency; misses must not panic
        let _ = get_owned(&mut c, &format!("k{}", (i + 3) % 6));
        let _ = get_owned(&mut c, "never-exists");
    }
    // With 6 rotating keys and cap 4, at most 4 distinct keys can hit.
    let hits = (0..6)
        .filter(|i| get_owned(&mut c, &format!("k{i}")).is_some())
        .count();
    assert!(
        hits <= 4,
        "expected <=4 survivors with cap 4 rotating over 6 keys, got {hits}"
    );
}

#[test]
fn full_reversal_then_insert_evicts_correct_victim() {
    // cap 5: fill a,b,c,d,e. Touch in reverse e,d,c,b,a so LRU=e.
    // Insert f must evict e.
    let mut c = Cache::new(5);
    for k in ["a", "b", "c", "d", "e"] {
        c.put(k, &k.to_uppercase());
    }
    for k in ["e", "d", "c", "b", "a"] {
        assert_hit(&mut c, k, &k.to_uppercase());
    }
    c.put("f", "F");
    assert_miss(&mut c, "e");
    assert_hit(&mut c, "a", "A");
    assert_hit(&mut c, "b", "B");
    assert_hit(&mut c, "c", "C");
    assert_hit(&mut c, "d", "D");
    assert_hit(&mut c, "f", "F");
}

#[test]
fn eviction_order_chain_two_rounds() {
    let mut c = Cache::new(3);
    c.put("a", "1");
    c.put("b", "2");
    c.put("c", "3");
    c.put("d", "4"); // evicts a
    assert_miss(&mut c, "a");
    c.put("e", "5"); // evicts b
    assert_miss(&mut c, "b");
    assert_hit(&mut c, "c", "3");
    assert_hit(&mut c, "d", "4");
    assert_hit(&mut c, "e", "5");
}
