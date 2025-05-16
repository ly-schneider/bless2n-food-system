import { Role } from "./Role";

export interface Club {
  _id: string;
  name: string;
  code: string;
  role?: Role
}