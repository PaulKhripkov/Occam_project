## Implementation description
Currently, realisation provides basic functionality of gathering price information from providers and transforming those into bars with a "fair" price. Implemented only features to satisfy the task.

### Differences from the original task
- **TickerPrice** struct renamed to **Bar**.
- Code is divided into packages.

### How it works
1. Create new price **Providers**.
2. Get a **Subscription** on a ticker from your **Providers**. You'll get a struct which can stream all prices from all providers in a single channel.
3. If you want to get a "fair" price in a form of bars with predefined interval, wrap your **Subscription** with **IndexPrice** and just read from a **brand-new Subscription**.

### What is not implemented but should be
- Errors from providers are not handled, they are just printed out.
- Errors in data not handled too.
- We assume that channels are properly controlled, and there is no way to discard some broken channel by timeout for instance.
- It's not possible to add providers in an existing subscription nor remove them from one. The only way to remove provider is to close price channel and send an error in appropriate channel.
- If a price stream from providers is too dense, the main loop of IndexPrice's goroutine may would not work properly. Though tests showed there are no such problems even if 100 providers each produces one bar every microsecond. Anyway we can control such behaviour by monitoring the time when timer fires.
