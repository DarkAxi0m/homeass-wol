"use client";
import ServerTable from "./_components/Servers";
import React from "react";
import { api } from "~/trpc/react";
import ServerEditDialog from "./_components/ServerEdit";
import { Server } from "@prisma/client";

export default function Home() {
  let servers = api.server.getAll.useQuery();
  const {data:actions} =  api.server.actions.useQuery()
  const [editServer, setEditServer] = React.useState<Server | null>(null);
  let stop = api.server.stop.useMutation();
  let {mutate:updateserver, error:updateerror} = api.server.update.useMutation({
    onSettled: () => {
      setEditServer(null);
      servers.refetch();
    },
  });

  let deleteserver = api.server.delete.useMutation({
    onSettled: () => {
      servers.refetch();
    },
  });


  const handleStartStop = (id: number) => {
    console.log(`Start/Stop server with id: ${id}`);
    stop.mutate({ id });
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
      startTypeId: null,
      stopTypeId: null,
      checkTypeId:null,
     lastSeen: null,
     lastChecked: null,
     isRunning: false,
     settings: "{}"

    };
    setEditServer(s);
  };

  const handleSave = (updatedServer: Server) => {
    console.log("updatedServer", updatedServer);
    updatedServer.startTypeId = updatedServer.startTypeId && updatedServer.startTypeId > 0 ? updatedServer.startTypeId : null;
    updatedServer.stopTypeId = updatedServer.stopTypeId && updatedServer.stopTypeId > 0 ? updatedServer.stopTypeId : null;
    updatedServer.checkTypeId = updatedServer.checkTypeId && updatedServer.checkTypeId > 0 ? updatedServer.checkTypeId : null;

    updateserver(updatedServer);
  };

  const handleCancel = () => {
    setEditServer(null);
  };

  console.log('Error', updateerror)

  const fieldErrors = updateerror?.data?.zodError?.fieldErrors
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

      {editServer && actions &&  (
        <ServerEditDialog
          server={editServer}
          onSave={handleSave}
          onCancel={handleCancel}
          actions={actions}
        />
      )}


      { fieldErrors && (
        <div className="mt-4 p-2 border  text-red-500 bg-black rounded">
          Save Failed: 
          {Object.keys(fieldErrors).map(k=><li>{k}: {fieldErrors[k]}</li>)}
        </div>
      )}
    </div>
  );
}
