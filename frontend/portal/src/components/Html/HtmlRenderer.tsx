import React from "react";
import he from "he";
import ReactMarkdown from 'react-markdown';

interface HtmlRendererProps {
  htmlString: string;
}

interface MarkdownRendererProps {
  data: string;
}

const HtmlTitleRenderer: React.FC<HtmlRendererProps> = ({ htmlString }) => {
  const decodedHtml = he.decode(htmlString);

  return (<div style={{ all: "revert" }} dangerouslySetInnerHTML={{ __html: decodedHtml }} />);
};

const HtmlBodyRenderer = ({ htmlString }: { htmlString: string }) => {
  const iframeRef = React.useRef<HTMLIFrameElement>(null);
  const decodedHtml = he.decode(htmlString);

  React.useEffect(() => {
    const iframe = iframeRef.current;

    if (iframe) {
      const doc = iframe.contentDocument || iframe.contentWindow?.document;

      if (doc) {
        doc.open();
        doc.write(`
          <style>body { margin-left: 0px; margin-right: 0px; }</style>
          ${decodedHtml}
        `);
        doc.close();

        const onLoadHandler = () => {
          iframe.style.height = doc.body.scrollHeight + 'px';

          const links = doc.querySelectorAll('a');
          links.forEach((link) => {
            link.setAttribute('target', '_blank');
            link.setAttribute('rel', 'noopener noreferrer');
          });
        };

        setTimeout(onLoadHandler, 0);
      }
    }
  }, [decodedHtml]);

  // Use decodedHtml hash as key to force full iframe re-render
  const iframeKey = React.useMemo(() => {
    return `iframe-${decodedHtml.length}-${Date.now()}`; // or use a hash if needed
  }, [decodedHtml]);

  return (
    <iframe
      key={iframeKey}
      ref={iframeRef}
      style={{
        width: '100%',
        border: 'none',
        overflow: 'hidden',
        height: '1px',
      }}
      title="HTML Preview"
    />
  );
};

const MarkdownRenderer: React.FC<MarkdownRendererProps> = ({ data }) => {
  const processedData = data.replace(/\n/g, "\n\n");
  return (
    <ReactMarkdown
      components={{
        p: ({ ...props }) => (
          <p style={{ all: "revert" }} {...props} />
        ),
      }}
    >
      {processedData}
    </ReactMarkdown>
  );
}


export { HtmlTitleRenderer, HtmlBodyRenderer, MarkdownRenderer };
