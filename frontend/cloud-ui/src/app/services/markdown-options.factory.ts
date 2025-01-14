import {MarkedOptions, MarkedRenderer} from 'ngx-markdown';

export function markedOptionsFactory(): MarkedOptions {
  const renderer = new MarkedRenderer();

  renderer.heading = (h) => {
    const xs = 6 - h.depth;
    const textClass = h.depth <= 4 ? `text-${xs}xl` : h.depth === 5 ? 'text-xl' : 'text-lg';
    const fontClass = h.depth === 1 ? 'font-extrabold' : 'font-bold';
    return `<h${h.depth} class='${textClass} ${fontClass} my-${xs}'>${h.text}</h${h.depth}>`;
  };

  renderer.code = (code) => {
    return `
        <pre>
          <code class="text-gray-900 dark:text-gray-200 whitespace-pre-line">
              ${code.text}
          </code>
        </pre>`;
  };

  return {
    renderer: renderer,
  };
}
