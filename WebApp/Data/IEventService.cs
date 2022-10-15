namespace WebApp.Data;

public class IEventService
{
    private string iEventsCsv = "./events.csv";

    public async Task<List<IEvent>> GetIEventsAsync()
    {
        List<IEvent> iEvents = await Task.Run(() => File.ReadAllLines(iEventsCsv)
                                  .Select(v => IEvent.FromCsv(v))
                                  .ToList());
        return iEvents;
    }
}