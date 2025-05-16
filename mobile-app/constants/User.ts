export interface User {
  _id: string;
  firstName: string;
  lastName: string;
  absenceDays?: {
    monday: boolean;
    tuesday: boolean;
    wednesday: boolean;
    thursday: boolean;
    friday: boolean;
    saturday: boolean;
    sunday: boolean;
  };
  absencePeriods?: {
    _id: string;
    startDate: Date;
    endDate: Date;
  }[];
}