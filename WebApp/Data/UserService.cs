namespace WebApp.Data;

public class UserService
{
    private string usersCsv = "./users.csv";

    public async Task<List<User>> GetUsers()
    {
        List<User> users = await Task.Run(() => File.ReadAllLines(usersCsv)
                                  .Select(v => User.FromCsv(v))
                                  .ToList());
        return users;
    }
}