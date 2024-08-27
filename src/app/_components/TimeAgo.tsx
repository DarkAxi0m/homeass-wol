"use client";

import React, { useState, useEffect } from "react";

interface TimeAgoProps {
  dateTime: Date | string;
}

const TimeAgo: React.FC<TimeAgoProps> = ({ dateTime }) => {
  const [timeAgo, setTimeAgo] = useState<string>("");
  const [updateInterval, setUpdateInterval] = useState<number | null>(null); // Default to no interval

  useEffect(() => {
    const calculateTimeAgo = () => {
      const now = new Date();
      const time = typeof dateTime === "string" ? new Date(dateTime) : dateTime;
      const seconds = Math.floor((now.getTime() - time.getTime()) / 1000);

      let interval: number | undefined;

      if (seconds >= 31536000) {
        interval = Math.floor(seconds / 31536000);
        setTimeAgo(`${interval} year${interval > 1 ? "s" : ""} ago`);
        setUpdateInterval(null); // No interval for years
      } else if (seconds >= 2592000) {
        interval = Math.floor(seconds / 2592000);
        setTimeAgo(`${interval} month${interval > 1 ? "s" : ""} ago`);
        setUpdateInterval(null); // No interval for months
      } else if (seconds >= 86400) {
        interval = Math.floor(seconds / 86400);
        setTimeAgo(`${interval} day${interval > 1 ? "s" : ""} ago`);
        setUpdateInterval(null); // No interval for days
      } else if (seconds >= 3600) {
        interval = Math.floor(seconds / 3600);
        setTimeAgo(`${interval} hour${interval > 1 ? "s" : ""} ago`);
        setUpdateInterval(3600 * 1000); // Update in an hour
      } else if (seconds >= 60) {
        interval = Math.floor(seconds / 60);
        setTimeAgo(`${interval} minute${interval > 1 ? "s" : ""} ago`);
        setUpdateInterval(60 * 1000); // Update in a minute
      } else {
        setTimeAgo(`${seconds} second${seconds > 1 ? "s" : ""} ago`);
        setUpdateInterval(1000); // Update every second
      }
    };

    calculateTimeAgo();

    if (updateInterval !== null) {
      const intervalId = setInterval(calculateTimeAgo, updateInterval);
      return () => clearInterval(intervalId);
    }
  }, [dateTime, updateInterval]);

  return <span>{timeAgo}</span>;
};

export default TimeAgo;
