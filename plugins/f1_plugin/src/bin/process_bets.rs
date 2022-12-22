use f1_plugin::consts;
use f1_plugin::entities::{Bet, EntityManager, RaceResult, User};
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

    let manager = EntityManager::new(consts::PATH);

    let mut race_results = match manager.from_csv::<RaceResult>("race_results") {
        Ok(race_results) => race_results,
        Err(error) => {
            println!("{}", error);

            return;
        }
    };

    if race_results[0].is_processed() {
        println!("Bets have already been processed for the current event.");

        return;
    }

    let mut bets = match manager.from_csv::<Bet>("bets") {
        Ok(bets) => bets,
        Err(error) => {
            println!("{}", error);

            return;
        }
    };

    let mut users = match manager.from_csv::<User>("users") {
        Ok(users) => users,
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

    match manager.to_csv::<User>("users", users) {
        Ok(()) => (),
        Err(_) => {
            println!("Error storing user points.");

            return;
        }
    }

    match manager.to_csv::<Bet>("bets", bets) {
        Ok(()) => (),
        Err(_) => {
            println!("Error storing bet points.");

            return;
        }
    }

    race_results[0].processed = format!("{}", race_results[0].event);

    match manager.to_csv::<RaceResult>("race_results", race_results) {
        Ok(()) => (),
        Err(_) => {
            println!("Error storing last processed bet.");

            return;
        }
    }

    println!("Bets successfully processed.");
}
