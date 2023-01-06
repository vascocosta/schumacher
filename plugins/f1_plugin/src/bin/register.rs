use csv_db::DataBase;
use f1_plugin::consts;
use f1_plugin::entities::User;
use std::env;

fn show_usage() {
    println!("Usage: !register");
}

fn main() {
    let args: Vec<String> = env::args().collect();

    if args.len() != 2 {
        show_usage();

        return;
    }

    let user = User::new(
        String::from(&args[1]).to_lowercase(),
        String::from("Europe/Berlin"),
        0,
        String::from(""),
    );

    let db = DataBase::new(consts::PATH, None);

    match db.select("users", None) {
        Ok(users) => if let Some(users) = users {
            if user.is_user(&users) {
                println!("You are already registered.");

                return;
            }
        }
        Err(_) => {
            println!("Error getting users.");

            return;
        }
    };

    match db.insert("users", user) {
        Ok(()) => println!("You were successfully registered."),
        Err(_) => println!("Error registering user."),
    }
}
