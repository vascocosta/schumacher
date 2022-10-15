namespace WebApp.Data;

public class User
{
    public string Nick { get; set; } = String.Empty;
    public string Timezone { get; set; } = String.Empty;
    public int Points { get; set; } = 0;
    public string Notifications { get; set; } = String.Empty;

    public static User FromCsv(string csvLine)
    {
        string[] values = csvLine.Split(',');
        User user = new User();
        user.Nick = values[0];
        user.Timezone = values[1];
        user.Points = Convert.ToInt32(values[2]);
        user.Notifications = values[3];
        return user;
    }
}