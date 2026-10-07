#[derive(Debug)]
pub struct Entry {
    pub key: String,
    pub val: String,
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
        if let Some(idx) = self.enteries.iter().position(|entery| entery.key == key) {
            let entry = self.enteries.remove(idx);
            let val = entry.val.clone();

            self.enteries.insert(0, entry);

            return Some(val);
        }

        None
    }

    pub fn put(&mut self, key: &str, val: &str) {
        if self.capacity == 0 {
            return;
        }

        if let Some(idx) = self.enteries.iter().position(|entery| entery.key == key) {
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
