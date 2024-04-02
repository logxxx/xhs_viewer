<template>
  <van-config-provider theme="dark">
    <van-floating-bubble v-model:offset="bubble_offset" icon="arrow-double-right" @click='speedup' />
  <div id="main_page">

    <van-swipe v-if="show" vertical class="swipe_ctrl" @change="onChangeVideo2" ref="swipe" :show-indicators=true :loop=false >
      <van-swipe-item v-for="(video,idx) in videos" :id=getSwipeID(video)>

          <video
              :id="video.id"
              class="videoSource"
              loop="loop" controls="controls" autoplay="autoplay"
              webkit-playsinline="true" x-webkit-airplay="true" playsInline={true} x5-playsinline="true" x5-video-orientation="portraint"
          >
            <source :src="getFileURL(video)"/>
          </video>

      </van-swipe-item>
    </van-swipe>

  </div>

  <van-tabbar id="tabbar" @change="onChangeTabbar">
    <van-tabbar-item icon=""></van-tabbar-item>
    <van-tabbar-item icon="delete-o">删除</van-tabbar-item>
    <van-tabbar-item icon="home-o">普通</van-tabbar-item>
    <van-tabbar-item icon="like-o">绝了</van-tabbar-item>
    <van-tabbar-item icon="eye-o">我看</van-tabbar-item>
    <van-tabbar-item icon="friends-o">其他</van-tabbar-item>
  </van-tabbar>
  </van-config-provider>
</template>
<script>
import axios from 'axios';
import {showToast} from 'vant';

export default {
  name: 'Video2',
  props: {
    msg: String
  },
  data() {
    return {
      show: true,
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

    speedup: function(){
      if(this.videos.length <= 0) {
        showToast("no video to speedup")
        return
      }
      let video = this.videos[0]
      if(!video){
        console.log("speedup failed: no video")
        return
      }
      let videoDom = document.getElementById(video.id)
      if(!videoDom){
        console.log("speedup failed: no video dom:", video.id)
        return
      }
      videoDom.currentTime += 5
      videoDom.play()
      console.log("speedup succ idx=", 0)
    },

    onChangeTabbar: function(idx){

      let action = ""
      switch (idx){
        case 1:
          action = "delete"
              break
        case 2:
          action = "normal"
              break
        case 3:
          action = "best"
          break
        case 4:
          action = "mine"
              break
        case 5:
          action = "other"
          break
      }
      //showToast("onChangeTabbar:" + action)

      this.DoAct(action)

    },

    DoAct: function(action){

      if(this.videos.length<=0){
        showToast("no video to act")
        return
      }

      //showToast("call act. idx="+this.watchingVideoIdx+" act="+action)
      let reqURL = this.getHost()+"viewer/act?action="+action+"&name="+this.videos[0].name+"&id="+this.videos[0].id

      console.log("["+action+"]"+this.videos[0].name)
      axios.get(reqURL).then(resp=>{
        //document.getElementById(this.videos[0].id).remove()
        //showToast("["+action+"]"+this.videos[0].name)
      })

      var swipeID = this.getSwipeID(this.videos[0])
      console.log("remove swipeID:", swipeID)
      //document.getElementById(swipeID).remove()
      this.videos.splice(0,1)
      console.log("splice result:", this.videos)
      if(this.videos.length>0){
        var msg = "["+1+"/"+this.videos.length+"]"+this.videos[0].size+ " "+this.videos[0].name
        showToast(msg)
        document.getElementById(this.videos[0].id).play()
      }


      if(this.videos.length<=1){
        this.getVideos()
      }



    },

    getFileURL: function(video) {
      var resp= this.getHost()+"viewer/file?name="+video.name+"&id="+video.id
      //console.log("getFileURL resp:", resp)
      return resp
    },

    getHost: function() {
      return "http://127.0.0.1:9887/"
      //return ""
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

        if(withSwipe===true) {
          this.$refs.swipe.next()
        }

      }).catch((err)=>{
        showToast("get videos catch err:"+err)
        console.log("get videos catch err:",err)
      })
    },


    isVideoDom(ele) {
      return ele && ele.tagName && ele.tagName === 'VIDEO'
    },

    onChangeVideo:function(idx) {
      //showToast("onChangeVideo:idx="+idx)

      let videoDom

      let v = this.videos[idx]
      if (v) {
        var msg = "["+(idx+1)+"/"+this.videos.length+"]"+v.size+ " "+v.name
        showToast(msg)
        console.log("play:", msg)
        videoDom = document.getElementById(v.id)
        if (this.isVideoDom(videoDom) && !v.isPlaying) {
          v.isPlaying = true
          videoDom.play().catch((err) => {
            //showToast("play video catch err:"+err)
          })
        }
      }

      this.videos.forEach((v, vIdx) => {
        if (vIdx == idx) {
          return
        }
        if(!v.isPlaying) {
          return
        }
        videoDom = document.getElementById(v.id)
        if (this.isVideoDom(videoDom)) {
          v.isPlaying = false
          videoDom.pause()
          videoDom.load() //重新加载视频元素以终止缓冲
          console.log("pause idx=", vIdx,"name=",v.name)
        }
      })
    }
  }
}
</script>

<style>
#main_page {
  height: 100vh;
  background-color: black;
}
.videoSource{
  width: 100vw;
}

.swipe_ctrl{
  height: 100%;
}

#tabbar{
  position: absolute;
  z-index: 999;
}


</style>