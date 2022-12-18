use crate::traits::Csv;
use std::error::Error;
use std::fs::File;
use std::fs::OpenOptions;

#[derive(Debug, Eq, Ord, PartialEq, PartialOrd)]
pub struct User {
    pub nick: String,
    pub time_zone: String,
    pub points: u32,
    pub notifications: String,
}

impl User {
    pub fn new(nick: String) -> Self {
        Self {
            nick,
            time_zone: String::from("Europe/Berlin"),
            points: 0,
            notifications: String::from(""),
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

impl Csv for User {
    fn from_csv(
        mut rdr: csv::Reader<File>,
        mut entities: Vec<User>,
    ) -> Result<Vec<User>, Box<dyn Error>> {
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

            entities.push(user);
        }

        Ok(entities)
    }

    /*
    fn to_csv(mut wtr: csv::Writer<File>, entity: Self) -> Result<(), Box<dyn Error>> {
        wtr.write_record(&[
            entity.nick,
            entity.time_zone,
            entity.points.to_string(),
            entity.notifications,
        ])?;

        wtr.flush()?;
        
        Ok(())
    }
    */

    fn to_csv(mut wtr: csv::Writer<File>, entities: Vec<Self>) -> Result<(), Box<dyn Error>> {
        for entity in entities {
            wtr.write_record(&[
                entity.nick,
                entity.time_zone,
                entity.points.to_string(),
                entity.notifications,
            ])?;
        }

        wtr.flush()?;
        
        Ok(())
    }
}

pub struct EntityManager {
    path: String,
}

impl EntityManager {
    pub fn new(path: String) -> Self {
        Self { path }
    }

    pub fn from_csv<T: Csv>(&self) -> Result<Vec<T>, Box<dyn Error>> {
        let rdr = csv::ReaderBuilder::new()
            .has_headers(false)
            .from_path(&self.path)?;

        let entities = Vec::new();

        T::from_csv(rdr, entities)
    }

    /*
    pub fn to_csv<T: Csv>(&self, entity: T) -> Result<(), Box<dyn Error>> {
        let file = OpenOptions::new()
            .write(true)
            .append(true)
            .open(&self.path)
            .unwrap();

        let wtr = csv::Writer::from_writer(file);

        T::to_csv(wtr, entity)
    }
    */

    pub fn to_csv<T: Csv>(&self, entities: Vec<T>) -> Result<(), Box<dyn Error>> {
        let file = OpenOptions::new()
            .write(true)
            .open(&self.path)
            .unwrap();

        let wtr = csv::Writer::from_writer(file);

        T::to_csv(wtr, entities)
    }

}
