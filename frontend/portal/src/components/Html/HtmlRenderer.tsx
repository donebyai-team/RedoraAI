// import React from "react";
// import he from "he";

// interface HtmlRendererProps {
//   htmlString: string;
// }

// const HtmlRenderer: React.FC<HtmlRendererProps> = ({ htmlString }) => {
//   const decodedHtml = he.decode(htmlString);
//   console.log("###_debug_decodedHtml ", decodedHtml);

//   return (<div dangerouslySetInnerHTML={{ __html: decodedHtml }} />);
// };

// export default HtmlRenderer;

import { useRef, useEffect } from "react";
import he from "he";

const HtmlRenderer = ({ htmlString }: { htmlString: string }) => {
  const iframeRef = useRef<HTMLIFrameElement>(null);
  const decodedHtml = he.decode(htmlString);

  useEffect(() => {
    const doc = iframeRef.current?.contentDocument;
    if (doc) {
      doc.open();
      doc.write(decodedHtml);
      doc.close();
    }
  }, [decodedHtml]);

  return (
    <iframe
      ref={iframeRef}
      style={{
        width: '100%',
        border: 'none',
        minHeight: 'max-c',
      }}
      title="HTML Preview"
    />
  );
};

export default HtmlRenderer;