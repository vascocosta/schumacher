use csv_db::DataBase;
use f1_plugin::consts;
use f1_plugin::entities::User;
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

    let db = DataBase::new(consts::PATH, None);

    let mut users: Vec<User> = match db.select("users", None) {
        Ok(users) => users.unwrap_or_default(),
        Err(_) => {
            println!("Error getting users.");

            return;
        }
    };

    users.sort_by(|a, b| b.points.cmp(&a.points));

    let mut position = 1;
    let mut output: String = Default::default();

    for user in users {
        if user.points > 0 {
            let re = Regex::new(r"[^A-Za-z0-9]+").unwrap();
            let nick = re.replace_all(&user.nick, "").to_uppercase();

            output = format!("{}{}. {} {} | ", output, position, &nick[..3], user.points);
            position += 1;
        }
    }

    println!("{}", output.trim_end_matches(" | "));
}
