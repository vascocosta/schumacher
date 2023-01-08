use csv_db::DataBase;
use f1_plugin::consts;
use f1_plugin::entities::Driver;
use std::env;

fn show_usage() {
    println!("Usage: !points");
}
fn main() {
    let args: Vec<String> = env::args().collect();

    if args.len() != 2 {
        show_usage();

        return;
    }

    let db = DataBase::new(consts::PATH, None);

    match db.select::<Driver>("drivers", None) {
        Ok(drivers) => if let Some(drivers) = drivers {
            for (index, driver) in drivers.iter().enumerate() {
                if index != drivers.len() - 1 {
                    print!("{} | ", driver.code);
                }
                else {
                    println!("{}", driver.code);
                }
            }
        }
        Err(_) => {
            println!("Could not get drivers.");

            return;
        }
    }
}