import { init, track, identify, setUserId, Identify } from '@amplitude/analytics-browser';

let amplitudeInitialized = false;

export function initAmplitude(apiKey: string) {
    if (!amplitudeInitialized) {
        init(apiKey, { defaultTracking: true });
        amplitudeInitialized = true;
    }
}

export function logDailyVisit(customerId: string, productName: string, metadata: Record<string, any> = {}) {
    if (!amplitudeInitialized) {
        console.warn('Amplitude not initialized');
        return;
    }

    // Set the user ID
    setUserId(customerId);

    // Use Identify to set user properties
    const identifyObj = new Identify()
        .set('organization_id', customerId)
        .set('organization_name', productName);

    identify(identifyObj);

    // Log the visit event
    track('Daily Visit', {
        timestamp: new Date().toISOString(),
        ...metadata,
    });
    console.log("daily visit event sent")
}
