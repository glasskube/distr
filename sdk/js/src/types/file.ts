import {BaseModel} from './base';

export interface DistrFile extends BaseModel{
  contentType: string;
  data: string;
  fileName: string;
  fileSize: number;
}
