<template>
  <van-config-provider theme="dark">
  <!--<van-button style="position:absolute;z-index: 999" @click="switchPing">PING</van-button>-->
  <div id="main_page">

      <video
          id="only_video"
          v-if="this.videos.length>0"
          :src="getFileURL()"
          class="videoSource"
          loop="loop" autoplay="autoplay" controls="false"
          webkit-playsinline="true" x-webkit-airplay="true" playsInline={true} x5-playsinline="true" x5-video-orientation="portraint"
      >
      </video>

    <van-floating-bubble type="" v-model:offset="show_btn_offset" icon="eye" @click='switchShowOpt' />

    <van-floating-bubble v-model:offset="reload_btn_offset" icon="arrow-double-right"  @click='doReload'/>
    <van-floating-bubble v-model:offset="preview_btn_offset" icon="arrow-left"  @click='this.isPreview = !this.isPreview'/>

    <van-floating-bubble v-model:offset="delete_btn_offset" icon="delete"  @click='doAction("delete")'/>

    <div v-show="showAct" style="position:absolute;bottom:50px;display:flex;">
      <van-space direction="vertical" fill style="margin:0 10px">
        <van-button class="btn" type="success" @click='doAction("richang")'>日常</van-button>
        <van-button class="btn" type="warning" @click='doAction("best")'>绝了</van-button>
        <van-button class="btn" type="primary" @click='doAction("good")'>不错</van-button>
        <van-button class="btn" type="" @click='doAction("normal")'>还行</van-button>
        <van-button class="btn" icon="arrow-double-right" type="primary" round @click="speedup"></van-button>
      </van-space>
      <van-space direction="vertical" fill style="">
        <van-button class="btn" type="" @click='doAction("other")'>其他</van-button>
        <van-button class="btn" type="" @click='doAction("fabu_putong")'>发普</van-button>
        <van-button class="btn" type="warning" @click='doAction("fabu_nv")'>发女</van-button>
        <van-button class="btn" type="primary" @click='doAction("mine")'>我看</van-button>
        <van-button class="btn" type="success" @click='doAction("foot")'>海底</van-button>
      </van-space>
      <div style="display:flex;flex-direction:column-reverse; margin:0 10px ">
        <div style="color:white;background-color:black;">
          {{this.getCurrentVideoName()}}
        </div>

      </div>
    </div>
  </div>
  </van-config-provider>
</template>
<script>
import axios from "axios";
import {showToast} from "vant";
import {showFailToast} from "vant/lib/toast/function-call";

export default {
  name: 'Video',
  props: {
    msg: String
  },
  data() {
    return {
      total: 0,
      showAct: true,
      show_btn_offset: {x: 20, y: 420},
      delete_btn_offset: {x: 20, y: 300},
      reload_btn_offset: {x: 20, y: 20},
      preview_btn_offset: {x: 350, y:20},
      videos: [],
      nextToken: '',
      watchingVideoIdx: 0,
      limit: 5,
      pingIntervalID: null,
      pingID: 0,
      isPreview: true,
    }
  },
  mounted(){
    this.getVideos()
  },
  created(){

  },
  methods: {

    switchPing:function() {
      if(this.pingIntervalID) {
        console.log("clearInterval")
        clearInterval(this.pingIntervalID)
        this.pingIntervalID = null
      }else{
        console.log("startPingTimer")
        this.startPingTimer()
      }
    },

    startPingTimer() {
      this.pingIntervalID = setInterval(this.testPing, 1000)
    },

    stopPingTimer() {
      if(this.pingIntervalID) {
        clearInterval(this.pingIntervalID)
      }
    },

    testPing: function() {
      this.pingID++
      var reqURL = this.getHost()+ "viewer/test_stream/" + this.pingID
      axios.get(reqURL)
    },

    switchShowOpt: function() {
      this.showAct = !this.showAct
    },

    getCurrentVideoName: function() {
      let video = this.getCurrentVideo()
      if(!video){
        return
      }
      return this.nextToken + "/" + this.total + " " + video.size+" "+video.name
    },

    getCurrentVideo: function() {
      if(this.videos.length<=0){
        return
      }
      return this.videos[this.watchingVideoIdx]
    },

    getHost: function() {
      //return "http://192.168.50.47:9887/"
      return ""
    },

    getFileURL: function() {
      var video = this.getCurrentVideo()
      if(!video){
        return
      }
      var resp= this.getHost()+"viewer/file?name="+video.name+"&id="+video.id + "&is_preview=" + this.isPreview
      console.log("getFileURL resp:", resp)
      return resp
    },

    getVideos: function() {

      let reqURL = this.getHost()+"viewer/videos?limit="+this.limit

      if(this.nextToken != "") {
        reqURL += "&next_token=" + this.nextToken
      }
      //showToast("getVideos start.url:"+reqURL)
      axios.get(reqURL).then(resp=>{

        //console.log("withSwipe:", withSwipe)

        if(!resp.data) {
          console.log("get videos failed.")
          return
        }
        if(!resp.data.videos){
          showToast("empty video")
          return
        }
        //showToast("get "+ resp.data.videos.length + " videos. token="+this.nextToken)
        resp.data.videos.forEach(v=>{
          this.videos.push({id: v.id, name: v.name, size:v.size})
        })
        this.total = resp.data.total
        this.nextToken = resp.data.next_token


      }).catch((err)=>{
        showToast("get videos catch err:"+err)
        console.log("get videos catch err:",err)
      })
    },

    doReload: function(){
      this.apiReload()
      this.nextToken = ""
      this.getVideos()
    },

    apiReload: function() {
      let reqURL = this.getHost()+"viewer/reload_video"
      axios.get(reqURL)
    },

    doAction: function(action){

      if(this.videos.length<=0){
        showToast("no video to act")
        return
      }

      let video = this.getCurrentVideo()
      if(!video) {
        showToast("no video idx to act")
        return
      }

      //showToast("call act. idx="+this.watchingVideoIdx+" act="+action)
      let reqURL = this.getHost()+"viewer/act?action="+action+"&id="+video.id

      console.log("["+action+"]"+reqURL)
      axios.get(reqURL).then(resp=>{
        if(resp.data.err_msg){
          showFailToast(resp.data.err_msg)
        }
      })

      this.watchingVideoIdx++

      if(this.watchingVideoIdx>=this.videos.length-2) {
        this.getVideos()
      }

      let videoDom = document.getElementById("only_video")
      if(videoDom){
        //videoDom.currentTime += 5
        videoDom.play()
      }


    },

    speedup: function(){
      try{
      let videoDom = document.getElementById("only_video")
        if(videoDom){
          videoDom.currentTime += 5
          videoDom.play()
        }
      }catch{

      }
    },

  }
}
</script>

<style>
#main_page {
  height:100vh;
  display:flex;
  background-color: black;
}
.videoSource{
  width: 100vw;
}

.btn{
  width:100px;
}

.btn_del {
  width:200px;
  position: absolute;
  z-index: 999;
  right: 80px;
  bottom: 80px;
}

</style>