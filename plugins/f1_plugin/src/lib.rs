use std::error::Error;
use std::fs::OpenOptions;

pub const USERS_FILE: &str = "/home/gluon/var/irc/bots/Schumacher/users.csv";

#[derive(Debug, Eq, Ord, PartialEq, PartialOrd)]
pub struct User {
    pub nick: String,
    pub time_zone: String,
    pub points: u32,
    pub notifications: String,
}

impl User {
    pub fn new(nick: String, time_zone: String, points: u32, notifications: String) -> User {
        User {
            nick,
            time_zone,
            points,
            notifications,
        }
    }

    pub fn from_path(path: &str) -> Result<Vec<User>, Box<dyn Error>> {
        let mut rdr = csv::ReaderBuilder::new()
            .has_headers(false)
            .from_path(path)?;

        let mut users = Vec::new();

        for result in rdr.records() {
            let record = result?;

            let user = User {
                nick: String::from(&record[0]),
                time_zone: String::from(&record[1]),
                points: match String::from(&record[2]).trim().parse() {
                    Ok(points) => points,
                    Err(_) => 0,
                },
                notifications: String::from(&record[3]),
            };

            users.push(user);
        }

        Ok(users)
    }

    pub fn to_path(&self, path: &str) -> Result<(), Box<dyn Error>> {
        let file = OpenOptions::new()
            .write(true)
            .append(true)
            .open(path)
            .unwrap();

        let mut wtr = csv::Writer::from_writer(file);

        wtr.write_record(&[
            &self.nick,
            &self.time_zone,
            &self.points.to_string(),
            &self.notifications,
        ])?;
        wtr.flush()?;

        Ok(())
    }

    pub fn is_user(&self, users: Vec<User>) -> bool {
        for user in &users {
            if user.nick == self.nick {
                return true;
            }
        }

        false
    }
}
