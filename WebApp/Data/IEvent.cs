namespace WebApp.Data;

public class IEvent
{
    public string Category { get; set;} = String.Empty;
    public string Title { get; set;} = String.Empty;
    public string Description { get; set;} = String.Empty;
    public string Date { get; set;} = String.Empty;
    public string DateCEST { get; set;} = String.Empty;
    public string DateEST { get; set;} = String.Empty;
    public string Channel { get; set;} = String.Empty;
    public string Image { get; set;} = String.Empty;
    public string Mention { get; set;} = String.Empty;

    public static IEvent FromCsv(string csvLine)
    {
        string[] values = csvLine.Split(',');
        IEvent iEvent = new IEvent();
        iEvent.Category = values[0];
        iEvent.Title = values[1];
        iEvent.Description = values[2];
        iEvent.Date = values[3];
        iEvent.Channel = values[4];
        iEvent.Image = values[5];
        iEvent.Mention = values[6];
        return iEvent;
    }
}