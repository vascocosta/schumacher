use std::error::Error;
use std::fs::File;

pub trait Csv {
    fn from_csv(rdr: csv::Reader<File>, entities: Vec<Self>) -> Result<Vec<Self>, Box<dyn Error>>
    where
        Self: Sized;
    
    fn to_csv(wtr: csv::Writer<File>, entities: Vec<Self>) -> Result<(), Box<dyn Error>>
    where
        Self: Sized;
}
