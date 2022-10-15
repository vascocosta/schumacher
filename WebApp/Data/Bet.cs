namespace WebApp.Data;

public class Bet
{
    public string Race { get; set; } = String.Empty;
    public string Nick { get; set; } = String.Empty;
    public string Driver1 { get; set; } = String.Empty;
    public string Driver2 { get; set; } = String.Empty;
    public string Driver3 { get; set; } = String.Empty;
    public string Points { get; set; } = String.Empty;

    public static Bet FromCsv(string csvLine)
    {
        string[] values = csvLine.Split(',');
        Bet bet = new Bet();
        bet.Race = values[0];
        bet.Nick = values[1];
        bet.Driver1 = values[2];
        bet.Driver2 = values[3];
        bet.Driver3 = values[4];
        bet.Points = values[5];
        return bet;
    }
}