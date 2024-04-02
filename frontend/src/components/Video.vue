<template>
  <van-config-provider theme="dark">
  <div id="main_page">

      <video
          id="only_video"
          v-if="this.videos.length>0"
          :src="getFileURL()"
          class="videoSource"
          loop="loop" autoplay="autoplay" controls="controls"
          webkit-playsinline="true" x-webkit-airplay="true" playsInline={true} x5-playsinline="true" x5-video-orientation="portraint"
      >
      </video>

    <van-floating-bubble type="" v-model:offset="bubble_offset" icon="eye" @click='switchShowOpt' />

    <div v-show="showAct" style="position:absolute;bottom:50px;display:flex;">
      <van-space direction="vertical" fill style="margin:0 10px">
        <van-button class="btn" type="success" @click='doAction("richang")'>日常</van-button>
        <van-button class="btn" type="warning" @click='doAction("best")'>绝了</van-button>
        <van-button class="btn" type="primary" @click='doAction("good")'>不错</van-button>
        <van-button class="btn" type="" @click='doAction("normal")'>普通</van-button>
        <van-button class="btn" icon="arrow-double-right" type="primary" round @click="speedup"></van-button>
      </van-space>
      <van-space direction="vertical" fill style="">
        <van-button class="btn" type="" @click='doAction("fabu")'>发布</van-button>
        <van-button class="btn" type="" @click='doAction("other")'>其他</van-button>
        <van-button class="btn" type="primary" @click='doAction("mine")'>我看</van-button>
        <van-button class="btn" type="success" @click='doAction("foot")'>海底</van-button>
        <van-button class="btn" type="warning" @click='doAction("delete")'>删除</van-button>
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

export default {
  name: 'Video',
  props: {
    msg: String
  },
  data() {
    return {
      showAct: true,
      bubble_offset: {x: 20, y: 420},
      videos: [],
      nextToken: '',
      watchingVideoIdx: 0,
      limit: 5,
    }
  },
  mounted(){
    this.getVideos()
  },
  created(){

  },
  methods: {

    switchShowOpt: function() {
      this.showAct = !this.showAct
    },

    getCurrentVideoName: function() {
      let video = this.getCurrentVideo()
      if(!video){
        return
      }
      return video.size+" "+video.name
    },

    getCurrentVideo: function() {
      if(this.videos.length<=0){
        return
      }
      return this.videos[this.watchingVideoIdx]
    },

    getHost: function() {
      //return "http://127.0.0.1:9887/"
      return ""
    },

    getFileURL: function() {
      var video = this.getCurrentVideo()
      if(!video){
        return
      }
      var resp= this.getHost()+"viewer/file?name="+video.name+"&id="+video.id
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

        this.nextToken = resp.data.next_token


      }).catch((err)=>{
        showToast("get videos catch err:"+err)
        console.log("get videos catch err:",err)
      })
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

</style>