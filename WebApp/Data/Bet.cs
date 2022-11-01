namespace WebApp.Data;

public class Bet : IComparable<Bet>
{
    public string Race { get; set; } = String.Empty;
    public string Nick { get; set; } = String.Empty;
    public string Driver1 { get; set; } = String.Empty;
    public string Driver2 { get; set; } = String.Empty;
    public string Driver3 { get; set; } = String.Empty;
    public int Points { get; set; } = 0;

    public static Bet FromCsv(string csvLine)
    {
        string[] values = csvLine.Split(',');
        Bet bet = new Bet();
        bet.Race = values[0];
        bet.Nick = values[1];
        bet.Driver1 = values[2];
        bet.Driver2 = values[3];
        bet.Driver3 = values[4];
        bet.Points = Convert.ToInt32(values[5]);
        return bet;
    }

    public int CompareTo(Bet bet)
    {
        if (this.Points < bet.Points) return -1;
        if (this.Points == bet.Points) return 0;
        return 1;
    }
}