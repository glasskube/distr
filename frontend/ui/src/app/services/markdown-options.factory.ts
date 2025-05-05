import {MarkedOptions, MarkedRenderer} from 'ngx-markdown';
import {parseInline} from 'marked';

export function markedOptionsFactory(): MarkedOptions {
  const renderer = new MarkedRenderer();
  const opts: MarkedOptions = {renderer: renderer};

  renderer.link = (link) => {
    const text = parseInline(link.text, {...opts, async: false});
    return `<a href="${link.href}" target="_blank">${text}</a>`;
  };

  return opts;
}
