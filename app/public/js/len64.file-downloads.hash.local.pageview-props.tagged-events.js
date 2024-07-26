!function(){"use strict";var o=window.location,p=window.document,l=p.currentScript,s=l.getAttribute("data-api")||new URL(l.src).origin+"/api/event";function e(e,t){try{if("true"===window.localStorage.plausible_ignore)return n=t,(a="localStorage flag")&&console.warn("Ignoring Event: "+a),void(n&&n.callback&&n.callback())}catch(e){}var a={},n=(a.n=e,a.u=o.href,a.d=l.getAttribute("data-domain"),a.r=p.referrer||null,t&&t.meta&&(a.m=JSON.stringify(t.meta)),t&&t.props&&(a.p=t.props),l.getAttributeNames().filter(function(e){return"event-"===e.substring(0,6)})),r=a.p||{},i=(n.forEach(function(e){var t=e.replace("event-",""),e=l.getAttribute(e);r[t]=r[t]||e}),a.p=r,a.h=1,new XMLHttpRequest);i.open("POST",s,!0),i.setRequestHeader("Content-Type","text/plain"),i.send(JSON.stringify(a)),i.onreadystatechange=function(){4===i.readyState&&t&&t.callback&&t.callback({status:i.status})}}var t=window.plausible&&window.plausible.q||[];window.plausible=e;for(var a,n=0;n<t.length;n++)e.apply(this,t[n]);function r(){a=o.pathname,e("pageview")}function u(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}window.addEventListener("hashchange",r),"prerender"===p.visibilityState?p.addEventListener("visibilitychange",function(){a||"visible"!==p.visibilityState||r()}):r();var c=1;function i(e){var t,a,n,r;if("auxclick"!==e.type||e.button===c)return t=function(e){for(;e&&(void 0===e.tagName||!u(e)||!e.href);)e=e.parentNode;return e}(e.target),a=t&&t.href&&t.href.split("?")[0],!function e(t,a){if(!t||w<a)return!1;if(y(t))return!0;return e(t.parentNode,a+1)}(t,0)&&(n=a)&&(r=n.split(".").pop(),g.some(function(e){return e===r}))?f(e,t,{name:"File Download",props:{url:a}}):void 0}function f(e,t,a){var n,r=!1;function i(){r||(r=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?(n={props:a.props},plausible(a.name,n)):(n={props:a.props,callback:i},plausible(a.name,n),setTimeout(i,5e3),e.preventDefault())}p.addEventListener("click",i),p.addEventListener("auxclick",i);var d=["pdf","xlsx","docx","txt","rtf","csv","exe","key","pps","ppt","pptx","7z","pkg","rar","gz","zip","avi","mov","mp4","mpeg","wmv","midi","mp3","wav","wma","dmg"],v=l.getAttribute("file-types"),m=l.getAttribute("add-file-types"),g=v&&v.split(",")||m&&m.split(",").concat(d)||d;function b(e){var e=y(e)?e:e&&e.parentNode,t={name:null,props:{}},a=e&&e.classList;if(a)for(var n=0;n<a.length;n++){var r,i=a.item(n).match(/plausible-event-(.+)(=|--)(.+)/);i&&(r=i[1],i=i[3].replace(/\+/g," "),"name"==r.toLowerCase()?t.name=i:t.props[r]=i)}return t}var w=3;function h(e){if("auxclick"!==e.type||e.button===c){for(var t,a,n,r,i=e.target,o=0;o<=w&&i;o++){if((n=i)&&n.tagName&&"form"===n.tagName.toLowerCase())return;u(i)&&(t=i),y(i)&&(a=i),i=i.parentNode}a&&(r=b(a),t?(r.props.url=t.href,f(e,t,r)):((e={}).props=r.props,plausible(r.name,e)))}}function y(e){var t=e&&e.classList;if(t)for(var a=0;a<t.length;a++)if(t.item(a).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}p.addEventListener("submit",function(e){var t,a=e.target,n=b(a);function r(){t||(t=!0,a.submit())}n.name&&(e.preventDefault(),t=!1,setTimeout(r,5e3),e={props:n.props,callback:r},plausible(n.name,e))}),p.addEventListener("click",h),p.addEventListener("auxclick",h)}();