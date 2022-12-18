use crate::traits::Csv;
use std::error::Error;
use std::fs::File;

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
        mut entities: Vec<Self>,
    ) -> Result<Vec<Self>, Box<dyn Error>> {
        for result in rdr.records() {
            let record = result?;

            let user = Self {
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

pub struct Bet {
    event: String,
    nick: String,
    first: String,
    second: String,
    third: String,
    points: u32,
}

impl Bet {
    fn new(event: String, nick: String, first: String, second: String, third: String) -> Self {
        Self {
            event,
            nick,
            first,
            second,
            third,
            points: 0,
        }
    }
}

impl Csv for Bet {
    fn from_csv(
        mut rdr: csv::Reader<File>,
        mut entities: Vec<Self>,
    ) -> Result<Vec<Self>, Box<dyn Error>> {
        for result in rdr.records() {
            let record = result?;

            let bet = Self {
                event: String::from(&record[0]),
                nick: String::from(&record[1]),
                first: String::from(&record[2]),
                second: String::from(&record[3]),
                third: String::from(&record[4]),
                points: match String::from(&record[5]).trim().parse() {
                    Ok(points) => points,
                    Err(_) => 0,
                },
            };

            entities.push(bet);
        }

        Ok(entities)
    }

    fn to_csv(mut wtr: csv::Writer<File>, entities: Vec<Self>) -> Result<(), Box<dyn Error>> {
        for entity in entities {
            wtr.write_record(&[
                entity.event,
                entity.nick,
                entity.first,
                entity.second,
                entity.third,
                entity.points.to_string(),
            ])?;
        }

        wtr.flush()?;

        Ok(())
    }
}

pub struct Driver {
    number: u32,
    code: String,
    odds: u32,
}

impl Csv for Driver {
    fn from_csv(
        mut rdr: csv::Reader<File>,
        mut entities: Vec<Self>,
    ) -> Result<Vec<Self>, Box<dyn Error>> {
        for result in rdr.records() {
            let record = result?;

            let driver = Self {
                number: match String::from(&record[0]).trim().parse() {
                    Ok(number) => number,
                    Err(_) => 0,
                },
                code: String::from(&record[1]),
                odds: match String::from(&record[2]).trim().parse() {
                    Ok(odds) => odds,
                    Err(_) => 0,
                },
            };

            entities.push(driver);
        }

        Ok(entities)
    }

    fn to_csv(mut wtr: csv::Writer<File>, entities: Vec<Self>) -> Result<(), Box<dyn Error>> {
        for entity in entities {
            wtr.write_record(&[
                entity.number.to_string(),
                entity.code,
                entity.odds.to_string(),
            ])?;
        }

        wtr.flush()?;

        Ok(())
    }
}

pub struct RaceResult {
    event: String,
    first: String,
    second: String,
    third: String,
    processed: String,
}

impl Csv for RaceResult {
    fn from_csv(
        mut rdr: csv::Reader<File>,
        mut entities: Vec<Self>,
    ) -> Result<Vec<Self>, Box<dyn Error>> {
        for result in rdr.records() {
            let record = result?;

            let race_result = Self {
                event: String::from(&record[0]),
                first: String::from(&record[1]),
                second: String::from(&record[2]),
                third: String::from(&record[3]),
                processed: String::from(&record[4]),
            };

            entities.push(race_result);
        }

        Ok(entities)
    }

    fn to_csv(mut wtr: csv::Writer<File>, entities: Vec<Self>) -> Result<(), Box<dyn Error>> {
        for entity in entities {
            wtr.write_record(&[
                entity.event,
                entity.first,
                entity.second,
                entity.third,
                entity.processed,
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

    pub fn to_csv<T: Csv>(&self, entities: Vec<T>) -> Result<(), Box<dyn Error>> {
        let wtr = csv::WriterBuilder::new()
            .has_headers(false)
            .from_path(&self.path)?;

        T::to_csv(wtr, entities)
    }
}
