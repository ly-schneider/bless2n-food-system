import { Club } from "./Club";
import { User } from "./User";

export interface Match {
  _id: string;
  club: Club;
  enemyClub: {
    name: string;
    address: string;
  };
  isHomeGame: boolean;
  date?: Date;
  participants?: (User | string)[];
}