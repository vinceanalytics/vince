

// Schedules function call for execution every interval. interval is a duration
// string.
//
// A duration string is a possibly signed sequence of
// decimal numbers, each with optional fraction and a unit suffix,
// such as "300ms", "-1.5h" or "2h45m".
// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
export function schedule(interval: string, call: () => void) {
    //@ts-ignore
    __schedule__(interval, call);
}