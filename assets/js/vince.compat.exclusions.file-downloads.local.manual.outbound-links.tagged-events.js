!function(){"use strict";var e,t,n,u=window.location,s=window.document,l=s.getElementById("vince"),f=l.getAttribute("data-api")||(e=l.src.split("/"),t=e[0],n=e[2],t+"//"+n+"/api/event");function v(e){console.warn("Ignoring Event: "+e)}function r(e,t){try{if("true"===window.localStorage.vince_ignore)return v("localStorage flag")}catch(e){}var n=l&&l.getAttribute("data-include"),r=l&&l.getAttribute("data-exclude");if("pageview"===e){var a=!n||n&&n.split(",").some(o),i=r&&r.split(",").some(o);if(!a||i)return v("exclusion rule")}function o(e){return u.pathname.match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var c={};c.n=e,c.u=t&&t.u?t.u:u.href,c.d=l.getAttribute("data-domain"),c.r=s.referrer||null,c.w=window.innerWidth,t&&t.meta&&(c.m=JSON.stringify(t.meta)),t&&t.props&&(c.p=t.props);var p=new XMLHttpRequest;p.open("POST",f,!0),p.setRequestHeader("Content-Type","text/plain"),p.send(JSON.stringify(c)),p.onreadystatechange=function(){4===p.readyState&&t&&t.callback&&t.callback()}}var a=window.vince&&window.vince.q||[];window.vince=r;for(var i=0;i<a.length;i++)r.apply(this,a[i]);function c(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}var p=1;function o(e){if("auxclick"!==e.type||e.button===p){var t,n=function(e){for(;e&&(void 0===e.tagName||!c(e)||!e.href);)e=e.parentNode;return e}(e.target),r=n&&n.href&&n.href.split("?")[0];if(!function e(t,n){if(!t||b<n)return!1;if(x(t))return!0;return e(t.parentNode,n+1)}(n,0))return(t=n)&&t.href&&t.host&&t.host!==u.host?d(e,n,{name:"Outbound Link: Click",props:{url:n.href}}):function(e){if(!e)return!1;var t=e.split(".").pop();return h.some(function(e){return e===t})}(r)?d(e,n,{name:"File Download",props:{url:r}}):void 0}}function d(e,t,n){var r=!1;function a(){r||(r=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented){var n=!t.target||t.target.match(/^_(self|parent|top)$/i),r=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type;return n&&r}}(e,t)?vince(n.name,{props:n.props}):(vince(n.name,{props:n.props,callback:a}),setTimeout(a,5e3),e.preventDefault())}s.addEventListener("click",o),s.addEventListener("auxclick",o);var m=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma"],g=l.getAttribute("file-types"),w=l.getAttribute("add-file-types"),h=g&&g.split(",")||w&&w.split(",").concat(m)||m;function y(e){var t=x(e)?e:e&&e.parentNode,n={name:null,props:{}},r=t&&t.classList;if(!r)return n;for(var a=0;a<r.length;a++){var i,o,c=r.item(a).match(/vince-event-(.+)=(.+)/);c&&(i=c[1],o=c[2].replace(/\+/g," "),"name"===i.toLowerCase()?n.name=o:n.props[i]=o)}return n}var b=3;function k(e){if("auxclick"!==e.type||e.button===p){for(var t,n,r,a,i=e.target,o=0;o<=b&&i;o++){if((r=i)&&r.tagName&&"form"===r.tagName.toLowerCase())return;c(i)&&(t=i),x(i)&&(n=i),i=i.parentNode}n&&(a=y(n),t?(a.props.url=t.href,d(e,t,a)):vince(a.name,{props:a.props}))}}function x(e){var t=e&&e.classList;if(t)for(var n=0;n<t.length;n++)if(t.item(n).match(/vince-event-name=(.+)/))return 1}s.addEventListener("submit",function(e){var t,n=e.target,r=y(n);function a(){t||(t=!0,n.submit())}r.name&&(e.preventDefault(),t=!1,setTimeout(a,5e3),vince(r.name,{props:r.props,callback:a}))}),s.addEventListener("click",k),s.addEventListener("auxclick",k)}();