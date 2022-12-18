use f1_plugin::consts;
use f1_plugin::entities::{EntityManager, User};
use regex::Regex;
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

    let manager = EntityManager::new(String::from(consts::USERS_FILE));

    //let mut users = match User::from_path(consts::USERS_FILE) {
    let mut users = match manager.from_csv::<User>() {
        Ok(users) => users,
        Err(_) => {
            println!("Error getting users.");

            return;
        }
    };

    users.sort_by(|a, b| b.points.cmp(&a.points));

    let mut position = 1;

    for user in &users {
        if user.points > 0 {
            let re = Regex::new(r"[^A-Za-z0-9]+").unwrap();
            let nick = re.replace_all(&user.nick, "").to_uppercase();

            print!("{}. {} {} | ", position, &nick[..3], user.points);

            position += 1;
        }
    }
}
