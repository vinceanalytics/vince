!function(){"use strict";var o=window.location,p=window.document,l=p.currentScript,s=l.getAttribute("data-api")||new URL(l.src).origin+"/api/event";function u(e,t){e&&console.warn("Ignoring Event: "+e),t&&t.callback&&t.callback()}function e(e,t){try{if("true"===window.localStorage.plausible_ignore)return u("localStorage flag",t)}catch(e){}var a=l&&l.getAttribute("data-include"),r=l&&l.getAttribute("data-exclude");if("pageview"===e){a=!a||a.split(",").some(n),r=r&&r.split(",").some(n);if(!a||r)return u("exclusion rule",t)}function n(e){var t=o.pathname;return(t+=o.hash).match(new RegExp("^"+e.trim().replace(/\*\*/g,".*").replace(/([^\.])\*/g,"$1[^\\s/]*")+"/?$"))}var a={},i=(a.n=e,a.u=o.href,a.d=l.getAttribute("data-domain"),a.r=p.referrer||null,t&&t.meta&&(a.m=JSON.stringify(t.meta)),t&&t.props&&(a.p=t.props),a.h=1,new XMLHttpRequest);i.open("POST",s,!0),i.setRequestHeader("Content-Type","text/plain"),i.send(JSON.stringify(a)),i.onreadystatechange=function(){4===i.readyState&&t&&t.callback&&t.callback({status:i.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var a,r=0;r<t.length;r++)e.apply(this,t[r]);function n(){a=o.pathname,e("pageview")}function c(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}window.addEventListener("hashchange",n),"prerender"===p.visibilityState?p.addEventListener("visibilitychange",function(){a||"visible"!==p.visibilityState||n()}):n();var f=1;function i(e){var t,a,r,n;if("auxclick"!==e.type||e.button===f)return t=function(e){for(;e&&(void 0===e.tagName||!c(e)||!e.href);)e=e.parentNode;return e}(e.target),a=t&&t.href&&t.href.split("?")[0],!function e(t,a){if(!t||h<a)return!1;if(k(t))return!0;return e(t.parentNode,a+1)}(t,0)&&(r=a)&&(n=r.split(".").pop(),b.some(function(e){return e===n}))?d(e,t,{name:"File Download",props:{url:a}}):void 0}function d(e,t,a){var r,n=!1;function i(){n||(n=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?(r={props:a.props},plausible(a.name,r)):(r={props:a.props,callback:i},plausible(a.name,r),setTimeout(i,5e3),e.preventDefault())}p.addEventListener("click",i),p.addEventListener("auxclick",i);var m=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],v=l.getAttribute("file-types"),g=l.getAttribute("add-file-types"),b=v&&v.split(",")||g&&g.split(",").concat(m)||m;function w(e){var e=k(e)?e:e&&e.parentNode,t={name:null,props:{}},a=e&&e.classList;if(a)for(var r=0;r<a.length;r++){var n,i=a.item(r).match(/plausible-event-(.+)(=|--)(.+)/);i&&(n=i[1],i=i[3].replace(/\+/g," "),"name"==n.toLowerCase()?t.name=i:t.props[n]=i)}return t}var h=3;function y(e){if("auxclick"!==e.type||e.button===f){for(var t,a,r,n,i=e.target,o=0;o<=h&&i;o++){if((r=i)&&r.tagName&&"form"===r.tagName.toLowerCase())return;c(i)&&(t=i),k(i)&&(a=i),i=i.parentNode}a&&(n=w(a),t?(n.props.url=t.href,d(e,t,n)):((e={}).props=n.props,plausible(n.name,e)))}}function k(e){var t=e&&e.classList;if(t)for(var a=0;a<t.length;a++)if(t.item(a).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}p.addEventListener("submit",function(e){var t,a=e.target,r=w(a);function n(){t||(t=!0,a.submit())}r.name&&(e.preventDefault(),t=!1,setTimeout(n,5e3),e={props:r.props,callback:n},plausible(r.name,e))}),p.addEventListener("click",y),p.addEventListener("auxclick",y)}();