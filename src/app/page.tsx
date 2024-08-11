"use client";
import ServerTable from "./_components/Servers";
import React from "react";
import { api } from "~/trpc/react";
import ServerEditDialog from "./_components/ServerEdit";
import { Server } from "~/lib/server";

export default function Home() {
  let servers = api.server.getAll.useQuery();
  let updateserver = api.server.update.useMutation({
    onSettled: () => {
      servers.refetch();
    },
  });

  let deleteserver = api.server.delete.useMutation({
    onSettled: () => {
      servers.refetch();
    },
  });

  const [editServer, setEditServer] = React.useState<Server | null>(null);

  const handleStartStop = (id: number) => {
    console.log(`Start/Stop server with id: ${id}`);
  };

  const handleEdit = (id: number) => {
    const serverToEdit =
      servers.data && servers.data.find((server) => server.id === id);
    if (serverToEdit) {
      setEditServer(serverToEdit);
    }
  };

  const handleDelete = (id: number) => {
    deleteserver.mutate({ id: id });
  };

  const handleAdd = () => {
    let s: Server = {
      id: -1,
      name: "",
      host: "",
      type: "simple",
    };
    setEditServer(s);
  };

  const handleSave = (updatedServer: Server) => {
    updateserver.mutate(updatedServer);
    setEditServer(null);
  };

  const handleCancel = () => {
    setEditServer(null);
  };

  return (
    <div>
      <div className="flex">
        <h2 className="text-2xl font-bold flex-grow">Server List</h2>
        <button
          onClick={() => handleAdd()}
          className="mr-2 text-green-600 hover:text-green-800"
        >
          ➕
        </button>
      </div>

      {servers.isLoading && <h1>Loading</h1>}
      {servers.data && (
        <ServerTable
          servers={servers.data}
          onStartStop={handleStartStop}
          onEdit={handleEdit}
          onDelete={handleDelete}
        />
      )}

      {editServer && (
        <ServerEditDialog
          server={editServer}
          onSave={handleSave}
          onCancel={handleCancel}
        />
      )}
    </div>
  );
}
