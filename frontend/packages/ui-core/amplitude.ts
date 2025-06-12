import { init, track, identify, setUserId, Identify } from '@amplitude/analytics-browser';

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
