use csv_db::CsvRecord;

#[derive(Debug)]
pub struct User {
    pub nick: String,
    pub time_zone: String,
    pub points: u32,
    pub notifications: String,
}

impl User {
    pub fn new(nick: String, time_zone: String, points: u32, notifications: String) -> Self {
        Self {
            nick,
            time_zone,
            points,
            notifications,
        }
    }
    pub fn is_user(&self, users: &Vec<User>) -> bool {
        for user in users {
            if user.nick == self.nick {
                return true;
            }
        }

        false
    }
}

impl CsvRecord for User {
    fn from_fields(fields: &Vec<String>) -> Self {
        Self {
            nick: String::from(&fields[0]),
            time_zone: String::from(&fields[1]),
            points: fields[2].parse().unwrap(),
            notifications: String::from(&fields[3]),
        }
    }

    fn to_fields(&self) -> Vec<String> {
        vec![
            self.nick.clone(),
            self.time_zone.clone(),
            self.points.to_string(),
            self.notifications.clone(),
        ]
    }
}

#[derive(Debug)]
pub struct Bet {
    pub event: String,
    pub nick: String,
    pub first: String,
    pub second: String,
    pub third: String,
    pub fourth: String,
    pub points: u32,
}

impl CsvRecord for Bet {
    fn from_fields(fields: &Vec<String>) -> Self {
        Self {
            event: String::from(&fields[0]),
            nick: String::from(&fields[1]),
            first: String::from(&fields[2]),
            second: String::from(&fields[3]),
            third: String::from(&fields[4]),
            fourth: String::from(&fields[5]),
            points: fields[6].parse().unwrap(),
        }
    }

    fn to_fields(&self) -> Vec<String> {
        vec![
            self.event.clone(),
            self.nick.clone(),
            self.first.clone(),
            self.second.clone(),
            self.third.clone(),
            self.fourth.clone(),
            self.points.to_string(),
        ]
    }
}

pub struct RaceResult {
    pub event: String,
    pub first: String,
    pub second: String,
    pub third: String,
    pub fourth: String,
    pub processed: String,
}

impl RaceResult {
    pub fn is_processed(&self) -> bool {
        self.processed.to_lowercase() == self.event.to_lowercase()
    }
}

impl CsvRecord for RaceResult {
    fn from_fields(fields: &Vec<String>) -> Self {
        Self {
            event: String::from(&fields[0]),
            first: String::from(&fields[1]),
            second: String::from(&fields[2]),
            third: String::from(&fields[3]),
            fourth: String::from(&fields[4]),
            processed: String::from(&fields[5]),
        }
    }

    fn to_fields(&self) -> Vec<String> {
        vec![
            self.event.clone(),
            self.first.clone(),
            self.second.clone(),
            self.third.clone(),
            self.fourth.clone(),
            self.processed.to_string(),
        ]
    }
}
