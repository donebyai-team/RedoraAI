import { ConnectError } from "@connectrpc/connect";

export function getConnectError(err: any): string {
    console.error("RPC error:", err);
    let userMessage = "Something went wrong";

    if (err instanceof ConnectError) {
        // Strip out the [code] prefix and GRPC-style `desc = ` part
        const match = err.message.match(/desc = (.+)$/);
        if (match) {
            userMessage = match[1]; // Clean, user-friendly message
        } else {
            // fallback to raw message without `[code]` prefix
            userMessage = err.message.replace(/^\[\w+_?\w*\]\s*/, '');
        }
    } else if (err?.response?.data?.message) {
        userMessage = err.response.data.message;
    } else if (err?.message) {
        userMessage = err.message;
    }

    return userMessage;
}