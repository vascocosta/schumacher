use std::env;
use std::error::Error;
use std::fmt;

#[derive(Debug)]
pub enum F1PluginError {
    ArgsError,
    DbError,
}

impl std::error::Error for F1PluginError {}

impl fmt::Display for F1PluginError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match self {
            F1PluginError::ArgsError => write!(f, "Error processing environment argument."),
            F1PluginError::DbError => write!(f, "Error accessing database."),
        }
    }
}

pub fn parse_args() -> Result<String, Box<dyn Error>> {
    let args: Vec<String> = env::args().collect();

    match args.len() != 2 {
        false => Ok(args[1].to_string()),
        true => Err(Box::new(F1PluginError::ArgsError)),
    }
}
