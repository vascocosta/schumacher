use f1_plugin::User;
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
        String::from(&args[1]),
        String::from("Europe/Berlin"),
        0,
        String::from(""),
    );

    let users = match User::from_path(f1_plugin::USERS_FILE) {
        Ok(users) => users,
        Err(_) => {
            println!("Error getting users.");

            return;
        }
    };

    if user.is_user(users) {
        println!("You are already registered.");

        return;
    }

    match user.to_path(f1_plugin::USERS_FILE) {
        Ok(()) => println!("You were successfully registered."),
        Err(_) => println!("Error registering user."),
    }
}
