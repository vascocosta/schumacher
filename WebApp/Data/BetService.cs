namespace WebApp.Data;

public class BetService
{
    private string betsCsv = "./bets.csv";

    public async Task<List<Bet>> GetBetsAsync()
    {
        List<Bet> bets = await Task.Run(() => File.ReadAllLines(betsCsv)
                                  .Select(v => Bet.FromCsv(v))
                                  .ToList());
        return bets;
    }
}