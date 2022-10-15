using System.Text.Json.Nodes;
namespace WebApp.Data;

public class ResultRow
{
    public string Position { get; set; } = String.Empty;
    public string Number { get; set; } = String.Empty;
    public string Driver { get; set; } = String.Empty;
}

public class QualifyingResultRow : ResultRow
{
    public string Q1 { get; set; } = String.Empty;
    public string Q2 { get; set; } = String.Empty;
    public string Q3 { get; set; } = String.Empty;
}

public class RaceResultRow : ResultRow
{
    public string FastestLap { get; set; } = String.Empty;
    public string Time { get; set; } = String.Empty;
}

public class ErgastF1Service
{
    private string? baseUrl;

    public ErgastF1Service()
    {
        baseUrl = "http://ergast.com/api/f1/current/last/";
    }

    public async Task<Tuple<string, List<QualifyingResultRow>>> QualifyingResults()
    {
        HttpClient client = new HttpClient();
        JsonNode? jsonNode = null;
        try
        {
            string responseBody = await client.GetStringAsync(baseUrl + "qualifying.json");
            jsonNode = JsonNode.Parse(responseBody);
        }
        catch
        {
            throw;
        }
        string raceName;
        try
        {
            raceName = jsonNode!["MRData"]!["RaceTable"]!["Races"]![0]!["raceName"]!.ToString();
        }
        catch (NullReferenceException)
        {
            raceName = String.Empty;
        }
        List<QualifyingResultRow> qualifyingResults = new List<QualifyingResultRow>();
        for (int i = 0; i < jsonNode!["MRData"]!["RaceTable"]!["Races"]![0]!["QualifyingResults"]!.AsArray().Count; i++)
        {
            QualifyingResultRow qualifyingResultRow = new();
            try
            {
                JsonNode? row = jsonNode!["MRData"]!["RaceTable"]!["Races"]![0]!["QualifyingResults"]![i];
                if (row != null)
                {
                    qualifyingResultRow.Position = row["position"]!.ToString();
                    qualifyingResultRow.Number = row["number"]!.ToString();
                    qualifyingResultRow.Driver = row["Driver"]!["code"]!.ToString();
                    qualifyingResultRow.Q1 = row["Q1"]!.ToString();
                    qualifyingResultRow.Q2 = row["Q2"]!.ToString();
                    qualifyingResultRow.Q3 = row["Q3"]!.ToString();
                }
            }
            catch (NullReferenceException)
            {
                // Some field was not available, so it defaults to String.Empty.
            }
            catch (ArgumentOutOfRangeException)
            {
                // Meh
            }
            qualifyingResults.Add(qualifyingResultRow);
        }
        return Tuple.Create(raceName, qualifyingResults);
    }

    public async Task<Tuple<string, List<RaceResultRow>>> RaceResults()
    {
        HttpClient client = new HttpClient();
        JsonNode? jsonNode = null;
        try
        {
            string responseBody = await client.GetStringAsync(baseUrl + "results.json");
            jsonNode = JsonNode.Parse(responseBody);
        }
        catch
        {
            throw;
        }
        string raceName;
        try
        {
            raceName = jsonNode!["MRData"]!["RaceTable"]!["Races"]![0]!["raceName"]!.ToString();
        }
        catch (NullReferenceException)
        {
            raceName = String.Empty;
        }
        List<RaceResultRow> raceResults = new ();
        for (int i = 0; i < jsonNode!["MRData"]!["RaceTable"]!["Races"]![0]!["Results"]!.AsArray().Count; i++)
        {
            RaceResultRow raceResultRow = new ();
            try
            {
                JsonNode row = jsonNode!["MRData"]!["RaceTable"]!["Races"]![0]!["Results"]![i]!;
                if (row != null)
                {
                    raceResultRow.Position = row["position"]!.ToString();
                    raceResultRow.Number = row["number"]!.ToString();
                    raceResultRow.Driver = row["Driver"]!["code"]!.ToString();
                    raceResultRow.FastestLap = row["FastestLap"]!["Time"]!["time"]!.ToString();
                    raceResultRow.Time = row["Time"]!["time"]!.ToString();
                }
            }
            catch (NullReferenceException)
            {
                // Some field was not available, so it defaults to String.Empty.
            }
            raceResults.Add(raceResultRow);
        }
        return Tuple.Create(raceName, raceResults);
    }
}