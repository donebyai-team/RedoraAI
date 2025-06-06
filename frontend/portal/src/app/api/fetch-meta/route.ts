import * as cheerio from 'cheerio';

export async function GET(request: Request): Promise<Response> {
    const { searchParams } = new URL(request.url);
    const url = searchParams.get('url');

    if (!url) {
        return new Response(JSON.stringify({ error: 'URL is required' }), {
            status: 400,
            headers: { 'Content-Type': 'application/json' },
        });
    }

    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 5000); // 5 seconds

    try {
        const res = await fetch(url, {
            headers: {
                'User-Agent':
                    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 ' +
                    '(KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36',
            },
            signal: controller.signal,
        });

        clearTimeout(timeout);

        const html = await res.text();
        const $ = cheerio.load(html);

        const title = $('title').text() || '';
        const description =
            $('meta[name="description"]').attr('content') ||
            $('meta[property="og:description"]').attr('content') ||
            '';

        return new Response(
            JSON.stringify({ title, description }),
            {
                status: 200,
                headers: { 'Content-Type': 'application/json' },
            }
        );
    } catch (error: any) {
        console.error('Fetch meta error:', error);

        const isTimeout = error.name === 'AbortError';

        return new Response(
            JSON.stringify({
                error: isTimeout ? 'Request timed out after 5 seconds' : 'Failed to fetch metadata',
            }),
            {
                status: 500,
                headers: { 'Content-Type': 'application/json' },
            }
        );
    }
}