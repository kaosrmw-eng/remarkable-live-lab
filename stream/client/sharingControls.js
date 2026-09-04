// Private controller: it never starts capture merely by opening the page.
let sharingKnown = false;
let sharingOn = false;
laserEnabled=false;trailEnabled=false;
async function refreshSharing() {
 const token=getAuthToken(); if(!token){document.getElementById('sharingState').textContent='Sign in to control sharing';return;}
 try {
  const response=await fetch('sharing',{headers:{Authorization:'Bearer '+token},cache:'no-store'});
  if(!response.ok)throw new Error('Sign in required');
  const state=await response.json();
  document.getElementById('sharingState').textContent=state.sharing?'● Sharing privately':state.reason;
  document.getElementById('startPrivateSharing').disabled=state.sharing;
  document.getElementById('stopPrivateSharing').disabled=!state.sharing;
  document.querySelectorAll('canvas').forEach(c=>c.style.visibility=state.sharing?'visible':'hidden');
  if(!state.sharing){
   if(reconnectTimeout){clearTimeout(reconnectTimeout);reconnectTimeout=null;}
   streamWorker.terminate(); eventWorker.terminate();
   hideReconnectBanner();hideConnectionError();
   document.getElementById('message').style.display='none';
   updateConnectionStatus(ConnectionState.DISCONNECTED);
  }
  if(sharingKnown&&!sharingOn&&state.sharing){location.reload();return;}
  sharingOn=state.sharing;sharingKnown=true;
 }catch(error){document.getElementById('sharingState').textContent='Offline / unavailable';document.querySelectorAll('canvas').forEach(c=>c.style.visibility='hidden');}
}
async function setSharing(action){
 try{
  const response=await fetch('sharing',{method:'POST',headers:{Authorization:'Bearer '+getAuthToken(),'Content-Type':'application/json'},body:JSON.stringify({action})});
  if(!response.ok)throw new Error('Control rejected');
  if(action==='start'){location.reload();return;} await refreshSharing();
 }catch(error){document.getElementById('sharingState').textContent=error.message;}
}
document.getElementById('startPrivateSharing').addEventListener('click',()=>setSharing('start'));
document.getElementById('stopPrivateSharing').addEventListener('click',()=>setSharing('stop'));
refreshSharing();setInterval(refreshSharing,1000);
