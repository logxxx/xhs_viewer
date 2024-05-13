import { createApp } from 'vue'
import Video from './components/Video.vue'
//import { Button } from 'vant-green';
//import { Col, Row } from 'vant-green';
//import { Swipe, SwipeItem } from 'vant-green';
import { Button } from 'vant';
import { Col, Row } from 'vant';
import { Swipe, SwipeItem } from 'vant';
import { Toast } from 'vant';
import 'vant/lib/index.css';
import "vant/es/toast/style"
import { FloatingBubble } from 'vant';
import { Tabbar, TabbarItem } from 'vant';
import { ConfigProvider } from 'vant';
import { Space } from 'vant';
import { Popup } from 'vant';
import { Field, CellGroup } from 'vant';
import { TextEllipsis } from 'vant';

const app = createApp(Video)
app.use(Button);
app.use(Col);
app.use(Row);
app.use(Toast);
app.use(FloatingBubble);
app.use(Swipe).use(SwipeItem);
app.use(Tabbar);
app.use(TabbarItem);
app.use(ConfigProvider);
app.use(Space);
app.use(Popup);
app.use(Field);
app.use(CellGroup);
app.use(TextEllipsis);


app.mount('#video_page')
