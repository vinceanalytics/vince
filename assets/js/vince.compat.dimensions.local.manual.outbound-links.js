!function(){"use strict";var e,t,n,u=window.location,o=window.document,c=o.getElementById("vince"),l=c.getAttribute("data-api")||(e=c.src.split("/"),t=e[0],n=e[2],t+"//"+n+"/api/event");function r(e,t){try{if("true"===window.localStorage.vince_ignore)return void console.warn("Ignoring Event: localStorage flag")}catch(e){}var n={};n.n=e,n.u=t&&t.u?t.u:u.href,n.d=c.getAttribute("data-domain"),n.r=o.referrer||null,n.w=window.innerWidth,t&&t.meta&&(n.m=JSON.stringify(t.meta)),t&&t.props&&(n.p=t.props);var r=c.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)}),a=n.p||{};r.forEach(function(e){var t=e.replace("event-",""),n=c.getAttribute(e);a[t]=a[t]||n}),n.p=a;var i=new XMLHttpRequest;i.open("POST",l,!0),i.setRequestHeader("Content-Type","text/plain"),i.send(JSON.stringify(n)),i.onreadystatechange=function(){4===i.readyState&&t&&t.callback&&t.callback()}}var a=window.vince&&window.vince.q||[];window.vince=r;for(var i=0;i<a.length;i++)r.apply(this,a[i]);var p=1;function s(e){if("auxclick"!==e.type||e.button===p){var t,n,r,a,i,o=function(e){for(;e&&(void 0===e.tagName||(!(t=e)||!t.tagName||"a"!==t.tagName.toLowerCase())||!e.href);)e=e.parentNode;var t;return e}(e.target);o&&o.href&&o.href.split("?")[0];if((i=o)&&i.href&&i.host&&i.host!==u.host)return t=e,r={name:"Outbound Link: Click",props:{url:(n=o).href}},a=!1,void(!function(e,t){if(!e.defaultPrevented){var n=!t.target||t.target.match(/^_(self|parent|top)$/i),r=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type;return n&&r}}(t,n)?vince(r.name,{props:r.props}):(vince(r.name,{props:r.props,callback:c}),setTimeout(c,5e3),t.preventDefault()))}function c(){a||(a=!0,window.location=n.href)}}o.addEventListener("click",s),o.addEventListener("auxclick",s)}();