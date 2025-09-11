import { init, track, identify, setUserId, Identify } from '@amplitude/analytics-browser';
import { setUser } from '@sentry/browser'

let amplitudeInitialized = false;

export function initAmplitude(apiKey?: string) {
    if (!apiKey) {
        console.log('Amplitude API key not found in env')
        return
    }

    if (!amplitudeInitialized) {
        init(apiKey, { defaultTracking: true });
        amplitudeInitialized = true;
        console.log('Amplitude initialized')
    }
}

export function logDailyVisit(customerId: string, productName: string, metadata: Record<string, any> = {}) {
    // Set user information in Decipher via the Sentry TypeScript SDK
    setUser({
        "id": customerId, // Optional: use if email not available
        "account": productName,  // Recommended: Which account/organization is this user a member of?
        "created_at": metadata.createdAt,
    });

    if (!amplitudeInitialized) {
        console.log('Amplitude not initialized');
        return;
    }

    // Set the user ID
    setUserId(customerId);

    // Use Identify to set user properties
    const identifyObj = new Identify()
        .set('organization_id', customerId)
        .set('organization_name', productName);

    identify(identifyObj);

    track('Daily Visit', {
        timestamp: new Date().toISOString(),
        ...metadata,
    }).promise.then((resp) => {
        console.log("Daily visit event sent successfully:", resp.message);
    }).catch((err) => {
        console.error("Error sending daily visit event:", err);
    });
}
