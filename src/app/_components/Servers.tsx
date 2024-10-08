"use client";
import React from "react";
import TimeAgo from "./TimeAgo";
import { Server } from "@prisma/client";

interface ServerTableProps {
  servers: Server[];
  onStartStop: (id: number) => void;
  onEdit: (id: number) => void;
  onDelete: (id: number) => void;
}

const ServerTable: React.FC<ServerTableProps> = ({
  servers,
  onStartStop,
  onEdit,
  onDelete,
}) => {
  return (
    <div className="overflow-x-auto">
      <table className="min-w-full bg-white rounded-lg shadow-sm">
        <thead>
          <tr>
            <th className="py-2 px-4 border-b text-left">Name</th>
            <th className="py-2 px-4 border-b text-left">Type</th>
            <th className="py-2 px-4 border-b text-left">Last Seen</th>
            <th className="py-2 px-4 border-b text-left">Last Checked</th>
            <th className="py-2 px-4 border-b text-left">Actions</th>
          </tr>
        </thead>
        <tbody>
          {servers.map((server) => (
            <tr key={server.id} className="hover:bg-gray-100">
              <td className="py-2 px-4 border-b">{server.name}</td>
              <td className="py-2 px-4 border-b">
                <small>
                  {server.startTypeId ? server.startTypeId + " " : ""}
                  {server.stopTypeId ? server.stopTypeId + " " : ""}
                  {server.checkTypeId ? server.checkTypeId : ""}
                </small>
                r
              </td>
              <td className="py-2 px-4 border-b">
                {server.lastSeen ? (
                  <TimeAgo dateTime={server.lastSeen} />
                ) : (
                  "N/A"
                )}
              </td>
              <td className="py-2 px-4 border-b">
                {server.lastChecked ? (
                  <TimeAgo dateTime={server.lastChecked} />
                ) : (
                  "N/A"
                )}
              </td>
              <td className="py-2 px-4 border-b">
                <button
                  onClick={() => onStartStop(server.id)}
                  className="mr-2 text-green-600 hover:text-green-800"
                >
                  🚀
                </button>
                <button
                  onClick={() => onEdit(server.id)}
                  className="mr-2 text-blue-600 hover:text-blue-800"
                >
                  ✏️
                </button>
                <button
                  onClick={() => onDelete(server.id)}
                  className="text-red-600 hover:text-red-800"
                >
                  🗑️
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default ServerTable;
