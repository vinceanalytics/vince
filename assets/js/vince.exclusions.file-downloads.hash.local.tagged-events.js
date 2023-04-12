!function(){"use strict";var u=window.location,s=window.document,l=s.currentScript,f=l.getAttribute("data-api")||new URL(l.src).origin+"/api/event";function v(e){console.warn("Ignoring Event: "+e)}function e(e,t){try{if("true"===window.localStorage.vince_ignore)return v("localStorage flag")}catch(e){}var n=l&&l.getAttribute("data-include"),r=l&&l.getAttribute("data-exclude");if("pageview"===e){var a=!n||n&&n.split(",").some(o),i=r&&r.split(",").some(o);if(!a||i)return v("exclusion rule")}function o(e){var t=u.pathname;return(t+=u.hash).match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var c={};c.n=e,c.u=u.href,c.d=l.getAttribute("data-domain"),c.r=s.referrer||null,c.w=window.innerWidth,t&&t.meta&&(c.m=JSON.stringify(t.meta)),t&&t.props&&(c.p=t.props),c.h=1;var p=new XMLHttpRequest;p.open("POST",f,!0),p.setRequestHeader("Content-Type","text/plain"),p.send(JSON.stringify(c)),p.onreadystatechange=function(){4===p.readyState&&t&&t.callback&&t.callback()}}var t=window.vince&&window.vince.q||[];window.vince=e;for(var n,r=0;r<t.length;r++)e.apply(this,t[r]);function a(){n=u.pathname,e("pageview")}function c(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}window.addEventListener("hashchange",a),"prerender"===s.visibilityState?s.addEventListener("visibilitychange",function(){n||"visible"!==s.visibilityState||a()}):a();var p=1;function i(e){if("auxclick"!==e.type||e.button===p){var t=function(e){for(;e&&(void 0===e.tagName||!c(e)||!e.href);)e=e.parentNode;return e}(e.target),n=t&&t.href&&t.href.split("?")[0];if(!function e(t,n){if(!t||y<n)return!1;if(x(t))return!0;return e(t.parentNode,n+1)}(t,0))return function(e){if(!e)return!1;var t=e.split(".").pop();return w.some(function(e){return e===t})}(n)?d(e,t,{name:"File Download",props:{url:n}}):void 0}}function d(e,t,n){var r=!1;function a(){r||(r=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented){var n=!t.target||t.target.match(/^_(self|parent|top)$/i),r=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type;return n&&r}}(e,t)?vince(n.name,{props:n.props}):(vince(n.name,{props:n.props,callback:a}),setTimeout(a,5e3),e.preventDefault())}s.addEventListener("click",i),s.addEventListener("auxclick",i);var o=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma"],m=l.getAttribute("file-types"),g=l.getAttribute("add-file-types"),w=m&&m.split(",")||g&&g.split(",").concat(o)||o;function h(e){var t=x(e)?e:e&&e.parentNode,n={name:null,props:{}},r=t&&t.classList;if(!r)return n;for(var a=0;a<r.length;a++){var i,o,c=r.item(a).match(/vince-event-(.+)=(.+)/);c&&(i=c[1],o=c[2].replace(/\+/g," "),"name"===i.toLowerCase()?n.name=o:n.props[i]=o)}return n}var y=3;function b(e){if("auxclick"!==e.type||e.button===p){for(var t,n,r,a,i=e.target,o=0;o<=y&&i;o++){if((r=i)&&r.tagName&&"form"===r.tagName.toLowerCase())return;c(i)&&(t=i),x(i)&&(n=i),i=i.parentNode}n&&(a=h(n),t?(a.props.url=t.href,d(e,t,a)):vince(a.name,{props:a.props}))}}function x(e){var t=e&&e.classList;if(t)for(var n=0;n<t.length;n++)if(t.item(n).match(/vince-event-name=(.+)/))return 1}s.addEventListener("submit",function(e){var t,n=e.target,r=h(n);function a(){t||(t=!0,n.submit())}r.name&&(e.preventDefault(),t=!1,setTimeout(a,5e3),vince(r.name,{props:r.props,callback:a}))}),s.addEventListener("click",b),s.addEventListener("auxclick",b)}();