import {MarkedOptions, MarkedRenderer} from 'ngx-markdown';
import {parseInline} from 'marked';
import {captureException} from '@sentry/angular';

export function markedOptionsFactory(): MarkedOptions {
  const renderer = new MarkedRenderer();
  const opts: MarkedOptions = {renderer: renderer};

  renderer.link = ({href, text}) => {
    try {
      text = text === href ? text : parseInline(text, {...opts, async: false});
    } catch (e) {
      captureException(e);
    }
    return `<a href="${href}" target="_blank">${text}</a>`;
  };

  return opts;
}
