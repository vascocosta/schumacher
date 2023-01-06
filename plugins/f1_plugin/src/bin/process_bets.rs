use csv_db::DataBase;
use f1_plugin::consts;
use f1_plugin::entities::{Bet, RaceResult, User};
use f1_plugin::utils;

const CORRECT: u32 = 5;
const PODIUM: u32 = 3;
const FASTEST_LAP: u32 = 1;
const BOOST: u32 = 10;

fn main() {
    let nick = match utils::parse_args() {
        Ok(nick) => nick,
        Err(error) => {
            println!("{}", error);

            return;
        }
    };

    if nick.to_lowercase() != "gluon" {
        println!("Only the bot admin can use this command.");

        return;
    }

    let db = DataBase::new(consts::PATH, None);

    let mut race_results: Vec<RaceResult> = match db.select("race_results", None) {
        Ok(race_results) => race_results.unwrap_or_default(),
        Err(error) => {
            println!("{}", error);

            return;
        }
    };

    if race_results[0].is_processed() {
        println!("Bets have already been processed for the current event.");

        return;
    }

    let mut bets: Vec<Bet> = match db.select("bets", None) {
        Ok(bets) => bets.unwrap_or_default(),
        Err(error) => {
            println!("{}", error);

            return;
        }
    };

    let mut users: Vec<User> = match db.select("users", None) {
        Ok(users) => users.unwrap_or_default(),
        Err(error) => {
            println!("{}", error);

            return;
        }
    };

    for bet in bets.iter_mut() {
        let mut score = 0;

        if bet.event.to_lowercase() == race_results[0].event.to_lowercase() {
            if [
                &race_results[0].first.to_lowercase(),
                &race_results[0].second.to_lowercase(),
                &race_results[0].third.to_lowercase(),
            ]
            .contains(&&bet.first.to_lowercase())
            {
                if bet.first.to_lowercase() == race_results[0].first.to_lowercase() {
                    score += CORRECT
                } else {
                    score += PODIUM
                }
            }

            if [
                &race_results[0].first.to_lowercase(),
                &race_results[0].second.to_lowercase(),
                &race_results[0].third.to_lowercase(),
            ]
            .contains(&&bet.second.to_lowercase())
            {
                if bet.second.to_lowercase() == race_results[0].second.to_lowercase() {
                    score += CORRECT
                } else {
                    score += PODIUM
                }
            }

            if [
                &race_results[0].first.to_lowercase(),
                &race_results[0].second.to_lowercase(),
                &race_results[0].third.to_lowercase(),
            ]
            .contains(&&bet.third.to_lowercase())
            {
                if bet.third.to_lowercase() == race_results[0].third.to_lowercase() {
                    score += CORRECT
                } else {
                    score += PODIUM
                }
            }

            if score == 3 * CORRECT {
                score += BOOST;
            }

            if bet.fourth.to_lowercase() == race_results[0].fourth.to_lowercase() {
                score += FASTEST_LAP;
            }

            bet.points = score;

            for user in users.iter_mut() {
                if user.nick.to_lowercase() == bet.nick.to_lowercase() {
                    user.points += score;
                }
            }
        }
    }

    match db.write("users", &users.iter().collect()) {
        Ok(()) => (),
        Err(_) => {
            println!("Error storing user points.");

            return;
        }
    }

    match db.write("bets", &bets.iter().collect()) {
        Ok(()) => (),
        Err(_) => {
            println!("Error storing bet points.");

            return;
        }
    }

    race_results[0].processed = format!("{}", race_results[0].event);

    match db.write("race_results", &race_results.iter().collect()) {
        Ok(()) => (),
        Err(_) => {
            println!("Error storing last processed bet.");

            return;
        }
    }

    println!("Bets successfully processed.");
}
