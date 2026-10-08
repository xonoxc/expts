#[derive(Debug)]
struct Entry {
    key: String,
    val: String,
}

#[derive(Debug)]
pub struct Cache {
    enteries: Vec<Entry>,
    capacity: usize,
}

impl Cache {
    pub fn new(cap: usize) -> Cache {
        Cache {
            enteries: vec![],
            capacity: cap,
        }
    }

    pub fn get(&mut self, key: &str) -> Option<String> {
        if let Some(idx) = self.enteries.iter_mut().position(|e| e.key == key) {
            let ent = self.enteries.remove(idx);
            let val = ent.val.clone();

            self.enteries.insert(0, ent);

            return Some(val);
        }

        None
    }

    pub fn put(&mut self, key: &str, val: &str) {
        if self.capacity == 0 {
            return;
        }

        if let Some(idx) = self.enteries.iter_mut().position(|e| e.key == key) {
            self.enteries.remove(idx);
        }

        self.enteries.insert(
            0,
            Entry {
                key: key.to_string(),
                val: val.to_string(),
            },
        );

        if self.enteries.len() > self.capacity {
            self.enteries.pop();
        }
    }
}
