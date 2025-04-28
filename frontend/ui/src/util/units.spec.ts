import {formatBytes} from './units';

describe('formatBytes', () => {
  it('should format 1000 to 1,000B', () => expect(formatBytes(1000, 'en-US')).toEqual('1,000B'));
  it('should format 1200 to 1.172KiB', () => expect(formatBytes(1200, 'en-US')).toEqual('1.172KiB'));
  it('should format -1024 to -1KiB', () => expect(formatBytes(-1024, 'en-US')).toEqual('-1KiB'));
  it('should format 8734568 to 8.330KiB', () => expect(formatBytes(1200, 'en-US')).toEqual('1.172KiB'));
  it('should format 1.5TiB to 1,536GiB', () => expect(formatBytes(1649267441664, 'en-US')).toEqual('1,536GiB'));
});
