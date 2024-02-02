(function(){"use strict";var n,s,o,i,a,r,c,g=window.location,t=window.document,e=t.currentScript,b=e.getAttribute("data-api")||u(e);function v(e){console.warn("Ignoring Event: "+e)}function u(e){return new URL(e.src).origin+"/api/event"}function l(n,s){try{if(window.localStorage.vince_ignore==="true")return v("localStorage flag")}catch{}var i,a,r,o={};o.n=n,o.u=s&&s.u?s.u:g.href,o.d=e.getAttribute("data-domain"),o.r=t.referrer||null,o.w=window.innerWidth,s&&s.meta&&(o.m=JSON.stringify(s.meta)),s&&s.props&&(o.p=s.props),r=e.getAttributeNames().filter(function(e){return e.substring(0,6)==="event-"}),a=o.p||{},r.forEach(function(t){var n=t.replace("event-",""),s=e.getAttribute(t);a[n]=a[n]||s}),o.p=a,o.h=1,i=new XMLHttpRequest,i.open("POST",b,!0),i.setRequestHeader("Content-Type","text/plain"),i.send(JSON.stringify(o)),i.onreadystatechange=function(){i.readyState===4&&s&&s.callback&&s.callback()}}i=window.vince&&window.vince.q||[],window.vince=l;for(n=0;n<i.length;n++)l.apply(this,i[n]);function p(e){for(;e&&(typeof e.tagName=="undefined"||!m(e)||!e.href);)e=e.parentNode;return e}function m(e){return e&&e.tagName&&e.tagName.toLowerCase()==="a"}function h(e,t){if(e.defaultPrevented)return!1;var n=!t.target||t.target.match(/^_(self|parent|top)$/i),s=!(e.ctrlKey||e.metaKey||e.shiftKey)&&e.type==="click";return n&&s}r=1;function d(e){if(e.type==="auxclick"&&e.button!==r)return;var t=p(e.target),n=t&&t.href&&t.href.split("?")[0];if(j(n))return f(e,t,{name:"File Download",props:{url:n}})}function f(e,t,n){var s=!1;function o(){s||(s=!0,window.location=t.href)}h(e,t)?(vince(n.name,{props:n.props,callback:o}),setTimeout(o,5e3),e.preventDefault()):vince(n.name,{props:n.props})}t.addEventListener("click",d),t.addEventListener("auxclick",d),a=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma"],o=e.getAttribute("file-types"),s=e.getAttribute("add-file-types"),c=o&&o.split(",")||s&&s.split(",").concat(a)||a;function j(e){if(!e)return!1;var t=e.split(".").pop();return c.some(function(e){return e===t})}})()