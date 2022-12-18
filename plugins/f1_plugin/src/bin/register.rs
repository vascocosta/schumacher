use f1_plugin::consts;
use f1_plugin::entities::{User, EntityManager};
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

    let user = User::new(String::from(&args[1]));

    let manager = EntityManager::new(String::from(consts::USERS_FILE));

    let mut users = match manager.from_csv::<User>() {
        Ok(users) => users,
        Err(_) => {
            println!("Error getting users.");

            return;
        }
    };

    if user.is_user(&users) {
        println!("You are already registered.");

        return;
    }

    users.push(user);

    match manager.to_csv::<User>(users) {
        Ok(()) => println!("You were successfully registered."),
        Err(_) => println!("Error registering user."),
    }
}
